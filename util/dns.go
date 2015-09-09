package util

import (
	"errors"
	"github.com/miekg/dns"
	"net"
)

var (
	/* dns server to use */
	DnsServer string

	/* default dns server */
	DefaultDnsServer = "8.8.8.8:53"
)

func get_dns_server() string {
	if DnsServer != "" {
		return DnsServer
	}

	return DefaultDnsServer
}

/*
Return all the ipv6 and ipv4 address for the domain name.

In the return list, ipv6 address is in front of ipv4 address.

if domain name resolve failed it will return an error.

Instead of system dns utils, we use the pure go dns library from https://github.com/miekg/dns

*/
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

/*
Return all the ipv4 address for the domain name.

If domain name resolve failed or get an emppty ip list will return an error.

Instead of system dns utils, we use the pure go dns library from https://github.com/miekg/dns

*/
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

/*
Return all the ipv6 address for the domain name.

If domain name resolve failed or get an emppty ip list will return an error.

Instead of system dns utils, we use the pure go dns library from https://github.com/miekg/dns

*/
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
