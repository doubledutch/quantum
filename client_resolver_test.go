package quantum

import "testing"

func TestNoAgentsErr(t *testing.T) {
	err := NoAgentsErr("test")
	if err.Error() != "no agents responded with type test" {
		t.Fatal("wrong error message")
	}
}

func TestNoAgentsWithNameErr(t *testing.T) {
	err := NoAgentsWithNameErr("test", "twice")
	if err.Error() != "no agents responded with type test and name twice" {
		t.Fatal("wrong error message")
	}
}

func TestIsNoAgents(t *testing.T) {
	errs := []error{NoAgentsWithNameErr("test", "twice"), NoAgentsErr("test")}

	for _, err := range errs {
		if !IsNoAgentsErr(err) {
			t.Fatal("expected agent err")
		}
	}
}

func TestNoAgentsFromRequest(t *testing.T) {
	r := ResolveRequest{
		Type: "test",
	}

	err := NoAgentsFromRequest(r)

	if err.Error() != "no agents responded with type test" {
		t.Fatal("wrong error message")
	}
}

func TestNoAgentsFromRequestWithAgent(t *testing.T) {
	r := ResolveRequest{
		Type:  "test",
		Agent: "twice",
	}

	err := NoAgentsFromRequest(r)

	if err.Error() != "no agents responded with type test and name twice" {
		t.Fatal("wrong error message")
	}
}
