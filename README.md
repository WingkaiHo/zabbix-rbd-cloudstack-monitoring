一.准备

1. 由于安装环境机器角色比较多，下面命令需要在什么情况，什么角色机器上安装说明

1） 例如下面命令需要在zabbix-server上面shell执行
    yum install zabbix zabbix-get zabbix-server-mysql zabbix-web-mysql zabbix-web zabbix-agent

    文档写为
    zabbix-server# yum install zabbix zabbix-get zabbix-server-mysql zabbix-web-mysql zabbix-web zabbix-agent 

2) 下面命命令需要在zabbix-agent执行. zabbix-agent就是代表所以被监控物理机器，都是zabbix-agent角色机器.
   zabbix-agent# yum install zabbix zabbix-agent

3) 下面的配置只需要在cloudstack management角色机器做命令写为:
   cloudstack-management# vim /etc/cloudstack/management/log4j-cloud.xml

2. Cloustack management服务监控需要在Cloustack源码对应模块打补丁以后才可以使用
   可以参考当前目录下 cloudstack/fix-cs-alarts-bug/ 相关说明.
   
3. Cloudstack management 模板监控事件说明
   参考当前目录下 cloudstack/doc-snmp-alert/
    
二. zabbix-server 安装

1. 添加 zabbix 官方/epel 源安装
    zabbix-server# rpm -ivh http://repo.zabbix.com/zabbix/2.4/rhel/7/x86_64/zabbix-release-2.4-1.el7.noarch.rpm

	zabbix-server# rpm -ivh http://dl.fedoraproject.org/pub/epel/7/x86_64/e/epel-release-7-5.noarch.rpm

2.  安装zabbix server
    由于server 同时需要监控所以这里也一并安装Zabbix-Agent:
    zabbix-server# yum install zabbix zabbix-get zabbix-server-mysql zabbix-web-mysql zabbix-web zabbix-agent

3. 安装mysql 和创建zabbix数据库
	1) 安装rpm包
		zabbix-server# yum install mariadb-server

	2) 修改配置问题
		zabbix-server# vim /etc/my.cnf
			[mysqld]
			....
			character-set-server=utf8
			innodb_file_per_table=1

	3）自动启动数据库
		zabbix-server# systemctl enable mariadb.service
		zabbix-server# systemctl start  mariadb.service

	3) 配置mysql默认密码
		zabbix-server# mysql_secure_installation

	5) 创建数据库

		zabbix-server# mysql -e "CREATE DATABASE zabbix character set utf8;" -u root -p
		zabbix-server# mysql -e "GRANT ALL PRIVILEGES ON zabbix.* TO 'zabbix'@'localhost' IDENTIFIED BY 'zabbix';" -u root -p
		zabbix-server# mysql -e 'FLUSH PRIVILEGES;' -u root -p 

		注意如果字符集不设置为UTF8在浏览器页面可能会乱码.

5. 导入数据到数据库
	zabbix-server# mysql -uzabbix -pzabbix
	
    MariaDB [zabbix]> use zabbix;
	MariaDB [zabbix]> source /usr/share/doc/zabbix-server-mysql-2.4.7/create/schema.sql

    需要注意如果安装zabbix-proxy 只需要导入schema.sql 就可以，无需导入下面的sql，否则zabbix-proxy 无法正常工作.

	MariaDB [zabbix]> source /usr/share/doc/zabbix-server-mysql-2.4.7/create/images.sql
	MariaDB [zabbix]> source /usr/share/doc/zabbix-server-mysql-2.4.7/create/data.sql

6. 配置zabbix_server.conf
	zabbix-server# egrep -v "(^#|^$)" /etc/zabbix/zabbix_server.conf

	修改前:
	LogFile=/var/log/zabbix/zabbix_server.log
	LogFileSize=0
	PidFile=/var/run/zabbix/zabbix_server.pid
	DBName=zabbix
	DBUser=zabbix
	DBSocket=/var/lib/mysql/mysql.sock
	SNMPTrapperFile=/var/log/snmptt/snmptt.log
	AlertScriptsPath=/usr/lib/zabbix/alertscripts
	ExternalScripts=/usr/lib/zabbix/externalscripts

	修改后:
	LogFile=/var/log/zabbix/zabbix_server.log
	LogFileSize=0
	PidFile=/var/run/zabbix/zabbix_server.pid
	DBHost=localhost							#配置mysql 服务hostname
	DBName=zabbix								#配置zabix mysql 数据库实例名字
	DBUser=zabbix								#配置访问数据库用户明
	DBPassword=zabbix							#配置访问数据库密码
	DBSocket=/var/lib/mysql/mysql.sock
	StartPollers=5
	SNMPTrapperFile=/var/log/snmptt/snmptt.log	
	CacheSize=256M								#配置cache 大小
	AlertScriptsPath=/usr/lib/zabbix/alertscripts
	ExternalScripts=/usr/lib/zabbix/externalscripts

	
7. 启动zabbix和http
	zabbix-server# service zabbix-server start
	zabbix-server# service httpd start 

	zabbix-server# systemctl enable zabbix-server.service
	zabbix-server# systemctl enable httpd.service

8. 防火墙配置
   zabbix-server# vim /ect/sysconfig/iptables

   -A INPUT -m state --state NEW -m tcp -p tcp --dport 22 -j ACCEPT
   -A INPUT -m state --state NEW -m tcp -p tcp --dport 80 -j ACCEPT
   -A INPUT -m state --state NEW -m tcp -p tcp --dport 10051 -j ACCEPT
   -A INPUT -m state --state NEW -m tcp -p tcp --dport 10050 -j ACCEPT
   
   1）10050是zabbix-agent端口，zabbix-agent采用被动方式，zabbix-server是主动连接agent的10050端口；
   2）10051是zabbix-serve端口， agent采用主动或trapper模式时候回访问server的端口;
	
  重新启动iptable服务:
  zabbix-server# service iptables restart

9. 关闭selinux
   1) 修改文件
   zabbix-server# vim /etc/selinux/config
	
   SELINUX=disabled

   2）执行下面命令
   zabbix-server# setenforce 0

10. 配置php.ini文件
	zabbix-server# vim /etc/php.ini

		date.timezone = Asia/Shanghai
		max_execution_time = 300
		post_max_size = 16M
		max_input_time = 300
		memory_limit = 128M

11. 初始化zabbix-server
	1)使用浏览器打开：
		 http://zabbix-server-ip/zabbix/
		 点击下一步进入
	2) Check of pre-requisites:
		如果各个配置都ok可以进入下一步。否则需要修改/etc/php.ini文件，并且重启http服务。

	3）Configure DB connection

		Database type: MySQL
		Database host:			#set the hostname of mysql
		Database port: 0        #0 use default port
		Database name: zabbix	
		Database user: zabbix
		Database password:		#database password in mysql
		
		然后点击Test connection 测试与数据库连接.连接成功，可以点击下一步

	4)	Zabbix server datails 
		host : localhost
		port:  10051
		name:  *****			#配置zabbix-server 名称。

	
		然后点击下一步，进入安装.

12. 服务器端mysql数据库zabbix历史事件表分区
    当有大量历史数据的时候数据数目查询，数据定时清理速度下降，提供zabbix效率建议对历史数据表进行分区.

1) 停止zabbix-server
   zabbix-server# service zabbix-server stop

2) 清理数据库
	zabbix-server# mysql -uzabbix -pzabbix

	mysql>use zabbix;
	mysql>truncate table history;
	mysql>optimize table history;
	mysql>truncate table history_str;
	mysql>optimize table history_str;
	mysql>truncate table history_uint;
	mysql>optimize table history_uint;
	mysql>truncate table trends;
	mysql>optimize table trends;
	mysql>truncate table trends_uint;
	mysql>optimize table trends_uint;
	mysql>truncate table events;
	mysql>optimize table events;

3) 使用网上开源zabbix分区脚本
	脚本在当前目录下 zabbix/partitiontables.sh. 把此脚本发送到zabbix-server机器上面执行.

4) 执行脚本
   为了防止执行期间网络中断导致脚本执行失败，使用screen把脚本在后台执行。如果没有screen请自行安装.

   zabbix-server# screen -R zabbix
   zabbix-server# bash partitiontables.sh

   按要求输入用户名:zabbix, 密码:zabbix, 选择备份数据库，连续按两次回车键. 当脚本用zabbix用户去连接数据库时候出现拒绝访问的提示，可以忽略直接按回车键即可.

   退出screen, 脚本将会在后台执行, 方法如下:
   按组合键CRTL+A之后再按组合键CRTL+D

   重新进入screen可以查看后台任务
   zabbix-server# screen -R zabbix

   注意: 严禁在脚本执行期间中断脚本运行，否则可能造成表损坏.   

三. zabbix-agent 安装

	1.zabbix 官方源安装
		zabbix-agent# rpm -ivh http://repo.zabbix.com/zabbix/2.4/rhel/7/x86_64/zabbix-release-2.4-1.el7.noarch.rpm

	2.epel 源
		zabbix-agent# rpm -ivh http://dl.fedoraproject.org/pub/epel/7/x86_64/e/epel-release-7-5.noarch.rpm

	3.安装zabbix-agent
		zabbix-agent# yum install zabbix zabbix-agent
		
		例如我们有1000台服务器需要监控，我们必须在1000台机器安装zabbix-agent, 通过zabbix-agent收集服务器性能数据，然后汇报到zabbix-server上。需要配合自动化安装，进行定制化参数配置。

	4. 防火墙配置
	
	zabbix-agent# vim /ect/sysconfig/iptables

		-A INPUT -m state --state NEW -m tcp -p tcp --dport 10051 -j ACCEPT
		-A OUTPUT -m state --state NEW -m tcp -p tcp --dport 10050 -j ACCEPT
   
		1）10050是zabbix-agent端口，zabbix-agent采用被动方式，zabbix-server是主动连接agent的10050端口；
		2）10051是zabbix-serve端口， agent采用主动或trapper模式时候回访问server的端口;
	
	
    重新启动iptable服务:
	zabbix-agent# service iptables restart

	5. 配置Agent
	zabbix-agent# vim /etc/zabbix/zabbix_agentd.conf

		Server=192.168.0.240
		ServerActive=192.168.0.240
		Hostname=				#设置名称与下一节你在zabbix-server添加监控host名称一致
		AllowRoot=1
		User=root

		1）被动模式配置:
			Server： 被动模式，允许哪台服务器连接Agent。

		2) 主动模式配置:
			ServerActive: 主动模式，向哪台服务器发送数据。

		3）允许多台服务器连接Agent，获取监控数据
			server=127.0.0.1,192.168.0.240 表示允许这两台服务器连接Agent

		4) 允许主动向多个server 发送监控数据
			ServerActive=192.168.0.240:10051

		5) 配置zabbix-agent 以root用户运行
			AllowRoot=1
			User=root
           如果不以root用户运行，所有被zabbix-agent执行脚本，访问log文件需要添加zabbix用户执行权限
		
	6. 配置 zabbix-agent开机自动启动
		zabbix-agent# systemctl enable zabbix-agent.service
		zabbix-agent# systemctl start  zabbix-agent.service


五. 安装graphite to zabbix 代理环境
    Cloudstack 资源数据, Ceph 监控数据存放在graphite时序数据库里面，zabbix可以通过下面方法获取.

1. 安装graphite_to_zabbxi以及相关的包
   1)首先自行安装python-pip

   2)把当前目录下 zabbix/graphite-to-zabbix目拷贝到zabbix-server上

   3)安装相关包
     zabbix-server# cd graphite-to-zabbix
     zabbix-server# yum install python-pyzabbix zabbix-sender
     zabbix-server# pip install ./configparser-3.3.0r2.tar.gz
     zabbix-server# pip install ./graphite-to-zabbix-0.1.6.tar.gz 
     zabbix-server# pip install ./py-zabbix-0.6.1.tar.gz

2. 添加定时任务
	
   定时获取graphite数据发送服务端
   zabbix-server# crontab -e

	输入:
	*/5 * * * * g2zproxy -z http://zabbix-server-ip/zabbix/ -zu admin -zp zabbix -g http://graphite-web-ip:port

   原理: zabbix item 的key 有这种模式模式graphite[......]都会通过g2zproxy从graphite抓取，并且通过zabbix-sender把结果发送zabbix-server。

六. 添加ceph集群监控

1. 导入ceph监控模板到zabbix-server
   ceph监控模块需要用到graphite to zabbix才可以使用，请参考步骤五进行安装. 
   ceph监控数据保存在graphite数据库，通过grafana-web 展示的。详细部署参考我另外项目https://github.com/WingkaiHo/ceph-monitor-dashboard
   
   1) 打开页面
      http://zabbix-server-ip/zabbix/

   2) 导入ceph集群监控模板
      点击configuration(配置)->Templates(模板) 然后group 选择ceph，点击Import
      按照要求选择xml文件(ceph模板xml 文件存放在当前目录下 zabbix/zbx_ceph_tmpl.xml)
      最后点击Import

2. 添加ceph监控集群主机

   ceph集群监控模板不能把它引用到普通主机上面，对应一个集群的状态，不是一个主机的状态.

   点击configuration(配置)->Host(主机) Group 选择ceph, 然后点击Create键按照下面要求输入:
   Host name: ceph_cluster               (必须填这个名称)
   Visible name:ceph_cluster             (必须填这个名称)
   Group In groups: ceph
   Agent interface: 把所有ceph mon ip都加入进去
   Monitored by proxy: (no proxy)
   Enabled true

   点击Templates tab
   link new templates 选择 Template Ceph Cluster and Graphite monnitor
   添加下面add

   点击最下方的add键

   添加成功后就可以对ceph集群进行监控.
   注意: 目前一套graphite，zabbix 环境只能监控一个ceph集群.

六. Cloudstack Management 服务监控.

   与ceph集群监控不同，一套zabbix环境可以监控多台cloudstack manager机器. 需要在zabbix-server机器安装snmp，snmptrap，配置host map

   1.把需要监控cloudstack management机器ip配置到hostmap
     zabbix-server# vim /etc/hosts  (添加机器名， ip对照表)
   ...  

   在hosts文件配置hostname和后面添到zabbix被监控机器的hostname必须是一致, 否则snmptt无法使用zabbix-sender发送状态到zabbix-server指定监控cloudstack manager 机器.

   2.安装snmp环境
     zabbix-server# yum install net-snmp net-snmp-utils
   
     1) 添加MIB文件到snmp服务
        把当前目录 cloudstack/CS-ROOT-MIB.txt 拷贝zabbix-server的/usr/share/snmp/mibs/目录, 告诉snmp加载我们mid

        zabbix-server# vim /etc/snmp/snmp.conf
		   mibs +CS-ROOT-MIB

     4) 自动启动snmp服务
        zabbix-server# systemctl enable snmpd.service
        zabbix-server# systemctl start  snmpd.service

     5) 测试MIB是否配置成功
        snmptranslate -On CS-ROOT-MIB::unallocatedVirtualNetworkpublicIp
   
   3.接受cloudstack发送snmptrap消息
     cloudstack management通过snmptrap 发送报警信息的，配置snmptrap server后才可以接受到报警.

     1) 安装snmptrap(接收报警信息服务)和snmptt(把cloudstack报警信息准化zabbix告警工具)环境
        zabbix-server# yum install net-snmp-perl perl-Config-IniFiles.noarch snmptt.noarch

     2）修改snmptrap 启动文件
        zabbix-server# vim /usr/lib/systemd/system/snmptrapd.service

        修改前:
        ExecStart=/usr/sbin/snmptrapd $OPTIONS -f 

        修改后: 
        ExecStart=/usr/sbin/snmptrapd $OPTIONS -On -f
        
        修改通过下面命令重新加载启动文件
        #systemctl daemon-reload

     3) 修改原因是:
        I. snmptt是基于数字OID来匹配/etc/snmp/snmptt.conf文件的内容，以确定接收到哪种陷入信息,并且把陷入信息标准化为对应的格式.
        II. 但是默认情况snnptrap把事件准化为:SNMPV2-MIB::coldStart，这种简化字符串snmptt无法与snmptt.conf匹配，所以无法识别, 因此需要修改脚本来启动解决.

     4)配置snmptrap 把接收命令传到snmptt处理
       zabbix-server# vim /etc/snmp/snmptrap.conf

       donotfork           yes
       printeventnumbers   yes
       traphandle          default /usr/sbin/snmptthandler
       ignoreauthfailure   yes
       authCommunity   log,execute,net public

     5)修改snmptt ini文件
       zabbix-server# vim /etc/snmp/snmptt.ini
       //修改项目
       mode = daemon
       date_time_format = %H:%M:%S %Y/%m/%d
       log_system_enable = 1
       daemon_uid = root
	   syslog_enable = 0 

     6)生成cloudstack转换zabbix告警对照表
       
       I.   snmptt conf文件可以通过下面命令生成snmpttconvertmib --in=/usr/share/snmp/mibs/CS-ROOT-MIB.txt --out=./snmptt.conf.CS-ROOT-MIB
       II.  配置把cloudstack对于告警发送到对应zabiix对应机器选项key值上
       III. conf文件已经配置好了。I，II 不需要做，只需把当前目录 cloudstack/CS-ROOT-MIB.txt 拷贝到zabbix-server机器/etc/snmp/, 重新启动
            snmp， snmptrap ，snmptt 服务

     7) 启动snmptrap
        zabbix-server# systemctl enable snmptrapd.service
        zabbix-server# systemctl restart snmptrapd.service    

     8) 启动snmptt
        zabbix-server# systemctl enable snmptt.service
        zabbix-server# systemctl start snmptt.service
 
	 9) 防火墙配置
        需要打开snmp， snmptrap端口
        zabbix-server# vim /ect/sysconfig/iptables
        -A INPUT -m state --state NEW -m udp -p udp --dport 161 -j ACCEPT
        -A INPUT -m state --state NEW -m udp -p udp --dport 162 -j ACCEPT
    
   4.Cloudstack Management 服务配置修改
      cloudstack-management# vim /etc/cloudstack/management/log4j-cloud.xml

      <appender name="SNMP" class="org.apache.cloudstack.alert.snmp.SnmpTrapAppender">
      <param name="Threshold" value="WARN"/>  <!-- Do not edit. The alert feature assumes WARN. -->
      <param name="SnmpManagerIpAddresses" value="10.1.1.1,10.1.1.2"/> <!-- 填入zabbix-server ip地址，可以填多个以','分割>
      <param name="SnmpManagerPorts" value="162,162"/> <!-- snmptrap 端口, 默认填入162就可以了>
      <param name="SnmpManagerCommunities" value="public,public"/>
      <layout class="org.apache.cloudstack.alert.snmp.SnmpEnhancedPatternLayout"> <!-- Do not edit -->
      <param name="PairDelimeter" value="//"/>
      <param name="KeyValueDelimeter" value="::"/>
      </layout>
      </appender>  

      修改以后:
      cloudstack-management# service cloudstack-management restart

   5.zabbix页面添加cloudstack management 模板
 
   1) 打开页面
      http://zabbix-server-ip/zabbix/ 
 
   2) 导入cloudstack management监控模板
      点击configuration(配置)->Templates(模板)->Import
      按照要求选择xml文件(cloudstack management模板xml文件存放在当前目录下 cloudstack/zbx_cs_templates.xml)
      最后点击Import

      完成以后就成功添加cloudstack management

   6.添加cloudstack management设备
 
     zabbix 监控以主机为单位，cloudstack management服务运行在一台主机上，监控cloudstack management状态先添加cloudstack management运行主机然后，主机temple 选择cloudstack management监控模板，就可以得到这台机器运行cloudstack management服务的报服务息. 一台机器可以连接多个模板.

   点击configuration(配置)->Host(主机) Group 选择Cloudstack Management, 然后点击Create键按照下面要求输入:
   Host name:                (必须填这个名称/etc/hosts ip map 对应名称一样)
   Visible name:             
   Group In groups: Cloudstack Management
   Agent interface:          (必须填这个名称/etc/hosts ip map hostname对于ip)
   Monitored by proxy: (no proxy)
   Enabled true

   点击Templates tab
   link new templates 选择 Template CloudStack Management
   添加下面add

   再点击最下方的add键

   重复上面的步骤添加多台的机器上cloudstack management服务状态进行监控。
