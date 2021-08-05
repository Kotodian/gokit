package utils

import (
	"errors"
	"net"
	"regexp"
)

var ipRexp *regexp.Regexp

func init() {
	ipRexp, _ = regexp.Compile("^[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}$")
}

func GetLocalIP() string {
	for {
		ifi, err := net.InterfaceByName("vboxnet0")
		if err != nil {
			break
		}
		addrs, err := ifi.Addrs()
		if err != nil {
			break
		}
		for _, a := range addrs {
			return a.(*net.IPNet).IP.String()
			// fmt.Printf("Interface %q, address %v\n", ifi.Name, a.Network())
		}
		break
	}

	// LocalIP:
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func IsPrivateIP(ip string) (bool, error) {
	if b := ipRexp.MatchString(ip); !b {
		return false, nil
	}
	var err error
	private := false
	//fmt.Println(ip)
	IP := net.ParseIP(ip)
	if IP == nil {
		err = errors.New("Invalid IP")
	} else {
		_, private24BitBlock, _ := net.ParseCIDR("10.0.0.0/8")
		_, private20BitBlock, _ := net.ParseCIDR("172.16.0.0/12")
		_, private16BitBlock, _ := net.ParseCIDR("192.168.0.0/16")
		private = private24BitBlock.Contains(IP) || private20BitBlock.Contains(IP) || private16BitBlock.Contains(IP)
	}
	return private, err
}
