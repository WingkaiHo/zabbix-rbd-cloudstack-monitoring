###1. Cloudstack Management Deemon Stop by Admin

    1) 执行启动cloudstack management 服务命令 trigger from PROBLEM to OK:
	cloudstack_manager#>service cloudstack-management start 
       执行命令以后
       CloudStack management status 状态是: Management server node $ip is up ($ip 代表cloudstack管理节点IP)

	2) 执行停止cloudstack management 服务命令 trigger from OK to PROBLEM
    cloudstack_manager#>service cloudstack-management stop (From OK to PROBLEM)
    
