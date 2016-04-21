// zabbix-add-interface project main.go
package main

import (
	"fmt"
	"flag"
	"regexp"
	"strings"
	"zabbix-add-interface/zabbix"
)

var user string
var passwd string
var url string
var agent_ip string
var host string
var port int 
var net_section string

func init() {
	
	flag.StringVar(&user, "user", "admin", "zabbix server user name.")
	flag.StringVar(&passwd, "password", "zabbix", "zabbix server user password.")
	flag.IntVar(&port, "port", 80, "The port of zabbix dashboard url.")
	flag.StringVar(&host, "host", "", "The host to add snmp idrac interface")
	flag.StringVar(&agent_ip, "agent_ip", "", "The zabbix agent ip address.")
	flag.StringVar(&net_section, "net_section", "", "The net secction of dell idrac")
	flag.Parse()
	
	if port != 80 {
		url = fmt.Sprintf("http://localhost:%d/zabbix/api_jsonrpc.php", port)
	} else {
		url = "http://localhost/zabbix/api_jsonrpc.php"
	}
}

// Find and return a single host object by name
func GetHost(api *zabbix.API, host string) (zabbix.ZabbixHost, error) {
    params := make(map[string]interface{}, 0)
    filter := make(map[string]string, 0)
    filter["host"] = host
    params["filter"] = filter
    params["output"] = "extend"
    params["select_groups"] = "extend"
    params["templated_hosts"] = 1
    ret, err := api.Host("get", params)

    // This happens if there was an RPC error
    if err != nil {
		var r zabbix.ZabbixHost
        return r, err
    }

    // If our call was successful
    if len(ret) > 0 {
        return ret[0], err
    }

    // This will be the case if the RPC call was successful, but
    // Zabbix had an issue with the data we passed.
	var r zabbix.ZabbixHost
    return r, &zabbix.ZabbixError{0,"","Host not found"}
}

func GetHostSnmpInterface(api *zabbix.API, hostid string)([]zabbix.HostInterface, error) {
	params := make(map[string]interface{}, 0)
    params["hostids"] = hostid
	filter := make(map[string]string, 0)
	filter["type"] = "2"
	params["filter"] = filter
    ret, err := api.Interface("get", params)
	
	return ret, err
}


func CreateHostSnmpInterface(api *zabbix.API, hostid, ip string) {
	params := make(map[string]interface{}, 0)
	params["hostid"] = hostid
	params["dns"] = ""
	params["ip"] = ip
	params["main"] = "1"
	params["port"] = "161"
	params["type"] = "2"
	params["useip"] = "1"
	
	_, err := api.Interface("create", params)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	fmt.Println("Create the interface successful")
}

func IsIp(ip string) (bool){
	if m, _ := regexp.MatchString("^[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}$", ip); !m {
		return false
	}
	return true
}

func main() {
	
	fmt.Printf("host:%s agent_ip:%s net_section:%s \n", host, agent_ip, net_section)
	if host == "" {
		fmt.Println("The host name could no be empty!!")
		return
	}
	
	if agent_ip == "" {
		fmt.Println("The host name could no be empty!!")
		return
	}
	
	if net_section == "" {
		fmt.Println("The net_section could not be empty!!")
		return
	}
	
	if IsIp(agent_ip) == false {
		fmt.Println("agent_ip is invalid!!")
		return
	}
	
	if IsIp(net_section) == false {
		fmt.Println("net_section is invalid!!")
		return
	}
	
	// build the idrac snmp ip adderss 
	// such as the zabbix-agent ip 192.168.20.7 and snmp net section is 192.168.230.0
	// idrac snmp ip is 192.168.230.7
	a := strings.Split(agent_ip, ".")
	b := strings.Split(net_section, ".")
	
	if len(a) != 4 || len(b) != 4 {
		fmt.Println("net_section or agent ip iss invalid!!")
		return
	}
	
	snmp_ip := b[0]+"."+b[1]+"."+b[2]+"."+a[3]
		
	api, err := zabbix.NewAPI(url, user, passwd)
    if err != nil {
        fmt.Println(err)
        return
    }

    _, err = api.Login()
    if err != nil {
        fmt.Println(err)
        return
    }
	
	h, err := GetHost(api, host)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	interfaces, err := GetHostSnmpInterface(api, h.HostId)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	// create the snmp interface
	if len(interfaces) == 0 {
		CreateHostSnmpInterface(api, h.HostId, snmp_ip)
	} else {
		fmt.Println("The idrac snmp interface has been add!!")
	}
}
