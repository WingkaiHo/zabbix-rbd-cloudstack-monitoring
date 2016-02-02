#一. 概述
##1. 我们当前使用Cloudstack 4.5.1 版本作为基础进行修改,使用cloudstack snmp 报警接口把报警消息发送zabbix-server snmptrap 端口上。
##2. 当前版本snmp 存在bug需要修改才可以达到我们目的.

#二. 修改Cloudstack snmp 接口和第三方报警模块通讯的bug
##1. snmp alarm 发送消息的时候却少时间戳的字段导致接受端解析OID时候错误。
     通过修改

###1) 修改文件SnmpHelper.java下createPDU 函数，加入时间戳字段， 所有情况都发送 DATA_CENTER_ID POD_ID CLUSTER_ID等字段
	   	
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

###2) 你可以通过目录下SnmpHelper.java替换或合并到你cs项目对应的文件.


##2. 修改里面managmentNode消息bug
###1). 此消息定义CS-ROOT-MIB.mib csAlertTraps 15 按照源码，management 节点启动和关闭时候都会发送trap消息。
###2). 服务开启的时候能够收到Management server node $ip is up消息,Management关闭以后没法收到Management server node $ip is down消息
###3). 通过打印ClusterManagerImpl.start调用堆栈,发现相关类缺少节点关闭逻辑导致无法执行CloudStackExtendedLifeCycle stop接口逻辑
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
		org.apache.cloudstack.spring.module.web.CloudStackContextLoaderListener.contextInitialized(CloudStackContextLoaderListener.java:52 需要在此类覆盖方法ContextLoaderListener.contextDestroyed(ServletContextEvent event)方法.

##4)serverlet 底下加载的moudule也没有执行stop方法顺便修改bug
        com.cloud.utils.component.ComponentContext (ComponentContext.java)添加uninitComponentsLifeCycle
        com.cloud.servlet.CloudStartupServlet (CloudStartupServlet.java) 覆盖HttpServlet.destroy函数

##5)修改 
