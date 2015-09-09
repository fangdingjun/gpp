package util

import (
	"fmt"
	"log"
	"testing"
)

func ExampleResolveDns() {
	res, err := ResolveDns("mail.google.com")
	if err != nil {
		log.Fatal(err)
	}
	for _, ip := range res {
		fmt.Printf("%s\n", ip.String())
	}
	// Output:
	// 2404:6800:4005:80a::2005
	// 74.125.203.17
	// 74.125.203.19
	// 74.125.203.18
	// 74.125.203.83
}

func ExampleResolveA() {
	res, err := ResolveA("mail.google.com")
	if err != nil {
		log.Fatal(err)
	}
	for _, ip := range res {
		fmt.Printf("%s\n", ip.String())
	}
	// Output:
	// 74.125.203.17
	// 74.125.203.19
	// 74.125.203.18
	// 74.125.203.83
}

func ExampleResolveAAAA() {
	res, err := ResolveAAAA("mail.google.com")
	if err != nil {
		log.Fatal(err)
	}
	for _, ip := range res {
		fmt.Printf("%s\n", ip.String())
	}
	// Output:
	// 2404:6800:4005:80a::2005
}

func TestResolveAAAA(t *testing.T) {
	res, err := ResolveAAAA("mail.google.com")
	if err != nil {
		t.Error(err)
	}

	if len(res) < 1 {
		t.Error("resolve a error")
	}

	for _, ip := range res {
		t.Logf("ip: %s\n", ip.String())
	}
}

func TestResolveA(t *testing.T) {
	res, err := ResolveA("mail.google.com")
	if err != nil {
		t.Error(err)
	}

	if len(res) < 1 {
		t.Error("resolve a error")
	}

	for _, ip := range res {
		t.Logf("ip: %s\n", ip.String())
	}
}

func TestResolveDns(t *testing.T) {
	res, err := ResolveDns("mail.google.com")
	if err != nil {
		t.Error(err)
	}

	if len(res) < 1 {
		t.Error("resolve a error")
	}

	for _, ip := range res {
		t.Logf("ip: %s\n", ip.String())
	}
}
