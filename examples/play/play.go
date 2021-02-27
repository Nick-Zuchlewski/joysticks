package main

import joysticks "github.com/Nick-Zuchlewski/joysticks"

func main() {
	evts, _ := joysticks.Capture(
		joysticks.Channel{Number: 1,
			Method: joysticks.HID.OnClose}, // event[0] chan set to receive button #1 closes events
	)
	<-evts[0]
}
