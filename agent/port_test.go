package agent

import "testing"

func TestPort(t *testing.T) {
	p := NewPort(":8500")

	if p.Value != ":8500" {
		t.Fatal("wrong port value")
	}

	if p.Int() != 8500 {
		t.Fatal("wrong port int")
	}
}
