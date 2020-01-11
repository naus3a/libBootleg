package libBootleg

import (
	"fmt"
	"net"
)

func GetLocalIps() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips
}

func GetOutboundIpInfo() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:8080")
	defer conn.Close()
	if err == nil {
		return conn.LocalAddr().(*net.UDPAddr).IP, err
	}
	return nil, err
}

func GetOutboundIp() string {
	/*var ip string
	ip = "127.0.0.1"
	conn, err := net.Dial("udp", "8.8.8.8:8080")
	if err == nil {
		ip = conn.LocalAddr().(*net.UDPAddr).IP.String()
	}
	if conn != nil {
		defer conn.Close()
	}
	return ip*/
	ip, err := GetOutboundIpInfo()
	if err != nil {
		return "127.0.0.1"
	} else {
		return ip.String()
	}

}

type NetInfo struct {
	Ip   string
	Port int
}

func (_ni NetInfo) String() string {
	return fmt.Sprintf("%v:%v", _ni.Ip, _ni.Port)
}
