package quantum

import "github.com/doubledutch/lager"

// UI handles sending logs to the client and agent
type UI interface {
	lager.Lager
	Client(log string)
	Both(log string)
}

// NewUI creates a new UI
func NewUI(conn AgentConn) UI {
	return &BasicUI{
		Lager:    conn.Lager(),
		clientCh: conn.Logs(),
	}
}

// BasicUI is a simple implementation of UI
type BasicUI struct {
	lager.Lager
	clientCh chan string
}

// Client logs to the client
func (ui *BasicUI) Client(log string) {
	ui.clientCh <- log
}

// Both logs to both the agent and the client
func (ui *BasicUI) Both(log string) {
	ui.Client(log)
	ui.Infof(log)
}
