#一. 概述
##1. 我们当前使用Cloudstack 4.5.1 版本作为基础进行修改,使用cloudstack snmp 报警接口把报警消息发送zabbix-server snmptrap(162)端口上。
##2. 当前版本snmp 存在bug需要修改才可以达到我们目的.

#二. 修改Cloudstack snmp 接口和第三方报警模块通讯的bug
##1. snmp alarm 发送消息的时候却少时间戳的字段导致接受端解析OID时候错误。
     通过修改

1) 修改文件SnmpHelper.java下createPDU 函数，加入时间戳字段， 所有情况都发送 DATA_CENTER_ID POD_ID CLUSTER_ID等字段
	   	
        if (alertType > 0) {
            trap.add(new VariableBinding(SnmpConstants.sysUpTime, new OctetString(new Date().toString())));
            trap.add(new VariableBinding(SnmpConstants.snmpTrapOID, getOID(CsSnmpConstants.TRAPS_PREFIX + alertType)));
            trap.add(new VariableBinding(getOID(CsSnmpConstants.DATA_CENTER_ID), new UnsignedInteger32(snmpTrapInfo.getDataCenterId())));
            trap.add(new VariableBinding(getOID(CsSnmpConstants.POD_ID), new UnsignedInteger32(snmpTrapInfo.getPodId())));
            trap.add(new VariableBinding(getOID(CsSnmpConstants.CLUSTER_ID), new UnsignedInteger32(snmpTrapInfo.getClusterId())));

            if (snmpTrapInfo.getMessage() != null) {
                trap.add(new VariableBinding(getOID(CsSnmpConstants.MESSAGE), new OctetString(snmpTrapInfo.getMessage())));
            } else {
                throw new CloudRuntimeException(" What is the use of alert without message ");
            }

            if (snmpTrapInfo.getGenerationTime() != null) {
                trap.add(new VariableBinding(getOID(CsSnmpConstants.GENERATION_TIME), new OctetString(snmpTrapInfo.getGenerationTime().toString())));
            } else {
                trap.add(new VariableBinding(getOID(CsSnmpConstants.GENERATION_TIME)));
            }
       }

2) 你可以通过目录下SnmpHelper.java替换或合并到你cs项目对应的文件.
3) 重新编译loud-plugin-snmp-alerts工程替换cloudstack cloud-plugin-snmp-alerts-4.5.1.jar.


##2. 修改里面managmentNode消息bug
1) 此消息定义CS-ROOT-MIB.mib csAlertTraps 15 按照源码，management 节点启动和关闭时候都会发送trap消息。
2) 服务开启的时候能够收到Management server node $ip is up消息,Management关闭以后没法收到Management server node $ip is down消息
3) 通过打印ClusterManagerImpl.start调用堆栈,代码如下:
            s_logger.info("Thrace ClusterManagerImpl::start() ");
            Map<Thread, StackTraceElement[]> stacks = Thread.getAllStackTraces();
            for (Map.Entry<Thread, StackTraceElement[]> entry : stacks.entrySet()) {
                Thread t = entry.getKey();
                s_logger.info("Thread :" + t.getName());
                StackTraceElement[] elems = entry.getValue();
                for (int i = 0; i < elems.length; i++) {
                    s_logger.info("\t" + elems[i].getClassName()+"."+elems[i].getMethodName()+"(" + elems[i].getFileName() + ":" + elems[i].getLineNumber() + ")");
                }
               s_logger.info("-------------------------------------");
            }
            s_logger.info("End the trace");
        }
4) 发现相关类缺少节点关闭逻辑导致无法执行CloudStackExtendedLifeCycle stop接口逻辑, start 接口调用堆栈如下:
		com.cloud.cluster.ClusterManagerImpl.start(ClusterManagerImpl.java:940)
		org.apache.cloudstack.spring.lifecycle.CloudStackExtendedLifeCycle$1.with(CloudStackExtendedLifeCycle.java:75)
		org.apache.cloudstack.spring.lifecycle.CloudStackExtendedLifeCycle.with(CloudStackExtendedLifeCycle.java:153)
		org.apache.cloudstack.spring.lifecycle.CloudStackExtendedLifeCycle.startBeans(CloudStackExtendedLifeCycle.java:72) 已经实现stop func 但是没有上层调用
		org.apache.cloudstack.spring.lifecycle.CloudStackExtendedLifeCycleStart.run(CloudStackExtendedLifeCycleStart.java:46) 
		org.apache.cloudstack.spring.module.model.impl.DefaultModuleDefinitionSet$1.with(DefaultModuleDefinitionSet.java:105)  需要在此类添加 stopContext stop 接口调用CloudStackExtendedLifeCycle停止逻辑
		org.apache.cloudstack.spring.module.model.impl.DefaultModuleDefinitionSet.withModule(DefaultModuleDefinitionSet.java:245)
		org.apache.cloudstack.spring.module.model.impl.DefaultModuleDefinitionSet.withModule(DefaultModuleDefinitionSet.java:250)
		org.apache.cloudstack.spring.module.model.impl.DefaultModuleDefinitionSet.withModule(DefaultModuleDefinitionSet.java:250)
		org.apache.cloudstack.spring.module.model.impl.DefaultModuleDefinitionSet.withModule(DefaultModuleDefinitionSet.java:233)
		org.apache.cloudstack.spring.module.model.impl.DefaultModuleDefinitionSet.startContexts(DefaultModuleDefinitionSet.java:97)
		org.apache.cloudstack.spring.module.model.impl.DefaultModuleDefinitionSet.load(DefaultModuleDefinitionSet.java:80)
		org.apache.cloudstack.spring.module.factory.ModuleBasedContextFactory.loadModules(ModuleBasedContextFactory.java:37) 需要在此类添加 stopModules 调用模块DefaultModuleDefinitionSet.stop 接口
		org.apache.cloudstack.spring.module.factory.CloudStackSpringContext.init(CloudStackSpringContext.java:70) 需要在此类添加 uinit接口 调研调用模块ModuleBasedContextFactory.stopModules接口
		org.apache.cloudstack.spring.module.factory.CloudStackSpringContext.<init>(CloudStackSpringContext.java:57)
		org.apache.cloudstack.spring.module.factory.CloudStackSpringContext.<init>(CloudStackSpringContext.java:61)
		org.apache.cloudstack.spring.module.web.CloudStackContextLoaderListener.contextInitialized(CloudStackContextLoaderListener.java:52)需要在此类覆盖方法ContextLoaderListener.contextDestroyed(ServletContextEvent event)方法.

5) 修改以后通过当前目录下: DefaultModuleDefinitionSet.java，ModuleBasedContextFactory.java，CloudStackSpringContext.java, CloudStackContextLoaderListener.java 替换原来工程文件,重新编译替换下面jar包: cloud-framework-cluster-4.5.1.jar, cloud-framework-spring-lifecycle-4.5.1.jar, cloud-framework-spring-module-4.5.1.jar.

##4 serverlet底下加载的moudule也没有执行关闭逻辑代码
1) 修改步骤如下:
        com.cloud.utils.component.ComponentContext (ComponentContext.java)添加uninitComponentsLifeCycle
        com.cloud.servlet.CloudStartupServlet (CloudStartupServlet.java) 覆盖HttpServlet.destroy函数
2) 可以通当前目录下ComponentContext.java, CloudStartupServlet.java替换原来文件, 替换cloud-utils-4.5.1.jar, cloud-server-4.5.1.jar

##5. 修改ClusterManagerImpl, 当服务节点关闭时候也执行节点离开的代码,以及ClusterAlertAdapter.onClusterNodeLeft

1) 修改ClusterManagerImpl.stop关闭方法
        if (_mshostId != null) {
            ManagementServerHostVO mshost = _mshostDao.findByMsid(_msId);
            //加入下面3句代码 
            List<ManagementServerHostVO> stopHostList = new ArrayList<ManagementServerHostVO>();
            stopHostList.add(mshost);
            notifyNodeLeft(stopHostList);
            //
            mshost.setState(ManagementServerHost.State.Down);
            _mshostDao.update(_mshostId, mshost);
            _mshostId=null;
        }
 
2) ClusterAlertAdapter onClusterNodeLeft,修改当节点退出的时候，如果退出节点是本机发送报警,非本机不发送报警

    90: if (mshost.getId() != args.getSelf().longValue()) 修改if (mshost.getId() == args.getSelf().longValue())

3) 把当前目录的ClusterManagerImpl.java ClusterAlertAdapter.java替换，然后编译.替换包cloud-framework-cluster-4.5.1.jar， cloud-server-4.5.1.jar


##6. 其他修改
1) 修改 server/src/com/cloud/network/router/VirtualNetworkApplianceManagerImpl.java的getRouterAlerts函数

   1471 修改后 _alertMgr.sendAlert(AlertType.ALERT_TYPE_DOMAIN_ROUTER, router.getDataCenterId(), router.getPodIdToDeployIn(), "Monitoring Service on VR " + router.getInstanceName() + alert, alert); 把alert信息打印到title上否则snmptrap无法收到alert信息
   修改后重新编译cloud-server-4.5.1.jar 

2) 修改 usage/src/com/cloud/usage/UsageManagerImpl.java SanityCheck类runInContext
1852 _alertMgr.sendAlert(AlertManager.AlertType.ALERT_TYPE_USAGE_SANITY_RESULT, 0, new Long(0), "Usage Sanity Check failed", errors);
   修改为
   _alertMgr.sendAlert(AlertManager.AlertType.ALERT_TYPE_USAGE_SANITY_RESULT, 0, new Long(0), "Usage Sanity Check failed"+errors, errors);
   修改后重新编译cloud-usage-4.5.1.jar

## 需要更新jar包统计
   cloud-plugin-snmp-alerts-4.5.1.jar
   cloud-framework-cluster-4.5.1.jar, 
   cloud-framework-spring-lifecycle-4.5.1.jar, 
   cloud-framework-spring-module-4.5.1.jar
   cloud-framework-cluster-4.5.1.jar
   cloud-utils-4.5.1.jar, 
   cloud-server-4.5.1.jar,
   cloud-usage-4.5.1.jar
