package quantum

import "log"

// UI handles sending logs to the client, the server (where it's used), or both
type UI interface {
	Log(log string)
	LogClient(log string)
	SetClient(client chan string)
	LogBoth(log string)
}

// UIConfig is configuration for UI
type UIConfig struct {
	System *log.Logger
	Client chan string
}

// DefaultUIConfig creates the default UI config
func DefaultUIConfig() *UIConfig {
	return new(UIConfig)
}

// NewUI creates a new UI
func NewUI(config *UIConfig) UI {
	if config == nil {
		config = DefaultUIConfig()
	}
	return NewBasicUI(config)
}

// BasicUI is a simple implementation of UI
type BasicUI struct {
	system *log.Logger
	client chan string
}

// NewBasicUI creates a new basic UI
func NewBasicUI(config *UIConfig) UI {
	return &BasicUI{
		system: config.System,
		client: config.Client,
	}
}

// LogClient logs to the client
func (ui *BasicUI) LogClient(log string) {
	ui.client <- log
}

// SetClient updates where the UI sends client logs
func (ui *BasicUI) SetClient(client chan string) {
	ui.client = client
}

// Log logs to the local logger
func (ui *BasicUI) Log(log string) {
	ui.system.Print(log)
}

// LogBoth logs both locally and the client
func (ui *BasicUI) LogBoth(log string) {
	ui.LogClient(log)
	ui.Log(log)
}
