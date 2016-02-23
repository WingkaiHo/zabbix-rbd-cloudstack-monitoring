一.准备

1. 由于安装环境机器角色比较多，下面命令需要在什么情况，什么角色机器上安装说明

1） 例如下面命令需要在zabbix-server上面shell执行
    yum install zabbix zabbix-get zabbix-server-mysql zabbix-web-mysql zabbix-web zabbix-agent

    文档写为
    zabbix-server# yum install zabbix zabbix-get zabbix-server-mysql zabbix-web-mysql zabbix-web zabbix-agent 

2) 下面命命令需要在zabbix-agent执行. zabbix-agent就是代表所以被监控物理机器，都是zabbix-agent角色机器.
   zabbix-agent# yum install zabbix zabbix-agent
    
二. zabbix-server 安装

1. 添加 zabbix 官方/epel 源安装
    zabbix-server# rpm -ivh http://repo.zabbix.com/zabbix/2.4/rhel/7/x86_64/zabbix-release-2.4-1.el7.noarch.rpm

	zabbix-server# rpm -ivh http://dl.fedoraproject.org/pub/epel/7/x86_64/e/epel-release-7-5.noarch.rpm

2.  安装zabbix server
    由于server 同时需要监控所以这里也一并安装Zabbix-Agent:
    zabbix-server# yum install zabbix zabbix-get zabbix-server-mysql zabbix-web-mysql zabbix-web zabbix-agent

3. 安装mysql 和创建zabbix数据库
	1) 安装rpm包
		zabbix-server# yum install mysql-server / yum install mariadb-server

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
	#mysql -uzabbix -pzabbix
	
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
	service iptables restart

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

   按要求输入用户名:zabbix, 密码为zabbix, 选择备份数据库，连续按两次回车键. 当脚本用zabbix用户去连接数据库时候出现拒绝访问的提示，可以忽略直接按回车键即可.

   退出screen,脚本将会在后台执行，方法如下
   按组合键CRTL+A之后再按组合键CRTL+D

   重新进入screen可以查看后台任务
   zabbix-server# screen -R zabbix

   ###注意: 严禁在脚本执行期间中断脚本运行，否则可能造成表损坏.   

三. zabbix-agent 安装

	1.zabbix 官方源安装
		zabbix-agent# rpm -ivh http://repo.zabbix.com/zabbix/2.4/rhel/7/x86_64/zabbix-release-2.4-1.el7.noarch.rpm

	2.epel 源
		zabbix-agent# rpm -ivh http://dl.fedoraproject.org/pub/epel/7/x86_64/e/epel-release-7-5.noarch.rpm

	3.安装zabbix-agent
		zabbix-agent# yum install zabbix zabbix-agent
		
		例如我们有1000台服务器需要监控，我们必须在1000台机器安装zabbix-agent,通过zabbix-agent收集服务器性能数据，然后汇报到zabbix-server上。需要配合自动化安装，进行定制化参数配置。

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

		Server=127.0.0.1
		ServerActive=127.0.0.1
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
     zabbix-server# pip install graphite_to_zabbix  

2. 添加定时任务
	
   定时获取graphite数据发送服务端
   zabbix-server# crontab -e

	输入:
	*/1 * * * * g2zproxy -z http://zabbix-server-ip/zabbix/ -zu admin -zp zabbix -g http://graphite-web-ip:port
