一.Ansible 安装
1. yum install ansible -y

二. Ansible 使用
   测试环境为3台机器， ansible-mgr: 192.168.20.200, host1: 192.168.20.201, host2: 192.168.20.202

	1. 配置DNS
	#vim /etc/hosts

	192.168.20.201 host1
	192.168.20.202 host2
	192.168.20.200 ansible-mgr-1

	2. add host1 and host2 know host
	#ssh-keyscan host1 host2 >> .ssh/known_hosts

	3. Hello World Ansible Style
	#ansible all -m ping --ask-pass
	
	host1 | SUCCESS => {
    "changed": false, 
    "ping": "pong"
	}
	host2 | SUCCESS => {
    "changed": false, 
    "ping": "pong"
	}


    4. 简单的playbook

	1) 添加collectd组
	   vim /etc/ansible/host
       [collectd]
	   host1
	   host2

	2) 编写exmple-play.yml
		然后执行
            ansible-playbook example.yml --ask-pass	


		
	 
