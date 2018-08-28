package connect

import (
	"net"
	"strconv"
	"fmt"
)

func ParseAddress(address string) (net.IP, int, error) {
	var strIp = ""
	var strPort = ""
	var port uint64 = 0
	var err error = nil

	if strIp, strPort, err = net.SplitHostPort(address); err == nil {
		if port, err = strconv.ParseUint(strPort, 10, 16); err == nil {
			if ip := net.ParseIP(strIp); ip != nil {
				return ip, int(port), nil
			}
			return nil, 0, fmt.Errorf("malformed IP address: %s", strIp)
		}
	}
	return nil, 0, err
}
