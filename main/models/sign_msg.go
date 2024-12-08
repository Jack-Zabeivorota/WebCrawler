package models

const SignalsTopic = "Signals"

type SignMsg struct {
	Sign     string         `json:"sign"`
	Services map[string]int `json:"services"`
}

var Signals = struct {
	Shutdown, Kill, ScaleUpdate string
}{
	Shutdown:    "shutdown",
	Kill:        "kill",
	ScaleUpdate: "scale_update",
}
