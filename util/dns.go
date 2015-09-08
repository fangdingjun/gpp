package util

import (
	"errors"
	"github.com/miekg/dns"
	"net"
)

var DnsServer string
var DefaultDnsServer = "8.8.8.8:53"

func get_dns_server() string {
	if DnsServer != "" {
		return DnsServer
	}

	return DefaultDnsServer
}

func ResolveDns(d string) ([]net.IP, error) {
	var data []net.IP
	res, err := ResolveAAAA(d)
	if err == nil && len(res) > 0 {
		data = append(data, res...)
	}
	res1, err := ResolveA(d)
	if err == nil && len(res1) > 0 {
		data = append(data, res1...)
	}
	if len(data) == 0 {
		return nil, errors.New("dns resolve failed")
	}
	return data, nil
}

func ResolveA(d string) ([]net.IP, error) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(d), dns.TypeA)
	m1, err := dns.Exchange(m, get_dns_server())
	if err != nil {
		return nil, err
	}
	if m1.Rcode != dns.RcodeSuccess {
		return nil, errors.New("dns resolve failed")
	}

	var res []net.IP
	for _, rr := range m1.Answer {
		if a, ok := rr.(*dns.A); ok {
			res = append(res, a.A)
		}
	}
	return res, nil
}

func ResolveAAAA(d string) ([]net.IP, error) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(d), dns.TypeAAAA)
	m1, err := dns.Exchange(m, get_dns_server())
	if err != nil {
		return nil, err
	}
	if m1.Rcode != dns.RcodeSuccess {
		return nil, errors.New("dns resolve failed")
	}

	var res []net.IP
	for _, rr := range m1.Answer {
		if a, ok := rr.(*dns.AAAA); ok {
			res = append(res, a.AAAA)
		}
	}
	return res, nil
}
