# joysticks
Go language joystick/controller/gamepad interface.

uses Linux 'input' interface to receive events directly, no polling.

uses channels to pipe around events, for flexibility and multi-threading.

Overview/docs: [![GoDoc](https://godoc.org/github.com/splace/joysticks?status.svg)](https://godoc.org/github.com/splace/joysticks)

Installation:

     go get github.com/splace/joysticks

Example: prints event info for pressing button #1 or moving hat#1.(with 10sec timeout.) 

	package main

	import . "github.com/splace/joysticks"
	import "fmt"
	import  "time"

	func main() {
		device := Connect(1)

		if device == nil {
			panic("no HIDs")
		}
		fmt.Printf("HID#1:- Buttons:%d, Hats:%d\n", len(device.Buttons), len(device.HatAxes)/2)

		// make channels for specific events
		b1press := device.OnClose(1)
		h1move := device.OnMove(1)

		// feed OS events onto the event channels. 
		go device.ParcelOutEvents()

		// handle event channels
		go func(){
			for{
				select {
				case <-b1press:
					fmt.Println("button #1 pressed")
				case h := <-h1move:
					hpos:=h.(HatPositionEvent)
					fmt.Println("hat #1 moved too:", hpos.X,hpos.Y)
				}
			}
		}()
	
		fmt.Println("Timeout in 10 secs.")
		<-time.After(time.Second*10)
		fmt.Println("Shutting down due to timeout.")
	}



Note: "jstest-gtk" - system wide mapping and calibration for joysticks.


