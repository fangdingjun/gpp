package util

import (
	"net"
	"time"
)

var DialTimeout time.Duration = 0
var DefaultDialTimeout time.Duration = 300 * time.Millisecond

func get_timeout() time.Duration {
	if DialTimeout != 0 {
		return DialTimeout
	}
	return DefaultDialTimeout
}

func Dial(network string, addr string) (net.Conn, error) {
	var ip net.IP
	var err error
	var ips []net.IP
	var conn net.Conn

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	/* test is ip address */
	ip = net.ParseIP(host)
	if ip == nil {
		/* domain name resolve */
		ips, err = ResolveDns(host)
		if err != nil {
			return nil, err
		}
	} else {
		ips = append(ips, ip)
	}

	for _, ip = range ips {
		conn, err = net.DialTimeout(network, net.JoinHostPort(ip.String(), port), get_timeout())
		if err == nil {
			/* dial success, return */
			return conn, nil
		}
		/* continue try next ip */
	}

	/* return last error */
	return nil, err
}
