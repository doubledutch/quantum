package consul

import (
	"net"
	"testing"

	"github.com/doubledutch/quantum"
	"github.com/miekg/dns"
)

func TestGetConfigs(t *testing.T) {
	expectedA, _ := net.ResolveTCPAddr("tcp", ":0")
	expectedTarget := "one.node.dc1.consul."
	expectedPort := uint16(1234)
	expectedStringPort := ":1234"

	msg := &dns.Msg{}
	msg.Answer = []dns.RR{
		&dns.SRV{
			Target: expectedTarget,
			Port:   expectedPort,
		},
	}
	msg.Extra = []dns.RR{
		&dns.A{
			A: expectedA.IP,
		},
	}

	results := newResolveResults(msg, quantum.ResolveRequest{
		Agent: "one",
		Type:  "",
	})

	if len(results) != 1 {
		t.Fatal("expected to resolve 1 config")
	}

	result := results[0]
	if result.address != expectedA.IP.String()+expectedStringPort {
		t.Fatal("wrong config addr")
	}
}

func TestBadTarget(t *testing.T) {
	expectedA, _ := net.ResolveTCPAddr("tcp", ":0")
	expectedTarget := "bad.dc1.consul."
	expectedPort := uint16(1234)

	msg := &dns.Msg{}
	msg.Answer = []dns.RR{
		&dns.SRV{
			Target: expectedTarget,
			Port:   expectedPort,
		},
	}
	msg.Extra = []dns.RR{
		&dns.A{
			A: expectedA.IP,
		},
	}

	results := newResolveResults(msg, quantum.ResolveRequest{
		Agent: "one",
		Type:  "",
	})
	if len(results) != 0 {
		t.Fatal("expected no results")
	}
}

func TestHostnameAgentDontMatch(t *testing.T) {
	expectedA, _ := net.ResolveTCPAddr("tcp", ":0")
	expectedTarget := "bad.node.dc1.consul."
	expectedPort := uint16(1234)

	msg := &dns.Msg{}
	msg.Answer = []dns.RR{
		&dns.SRV{
			Target: expectedTarget,
			Port:   expectedPort,
		},
	}
	msg.Extra = []dns.RR{
		&dns.A{
			A: expectedA.IP,
		},
	}

	results := newResolveResults(msg, quantum.ResolveRequest{
		Agent: "one",
		Type:  "",
	})

	if len(results) != 0 {
		t.Fatal("expected no results")
	}
}
