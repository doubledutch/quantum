package agent

import "strconv"

// Port holds a port in the form :XXXX
type Port struct {
	Value string
}

// NewPort returns a Port
func NewPort(s string) Port {
	return Port{Value: s}
}

// Int returns a int representation of Port
func (p Port) Int() int {
	number := p.Value[1:] // ignore : at [0]

	i, err := strconv.Atoi(number)
	if err != nil {
		return -1
	}
	return i
}
