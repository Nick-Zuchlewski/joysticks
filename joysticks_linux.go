// +build linux

package joysticks

import (
	"encoding/binary"
	"io"
	"os"
	"strconv"
	"time"
)

// see; https://www.kernel.org/doc/Documentation/input/joystick-api.txt
type osEventRecord struct {
	Time  uint32 // event timestamp, unknown base, in milliseconds 32bit, so about a month
	Value int16  // value
	Type  uint8  // event type
	Index uint8  // axis/button
}

const maxValue = 1<<15 - 1

// common path root, so Connect and DeviceExists are not thread safe.
var inputPathSlice = []byte("/dev/input/js ")[0:13]

// DeviceExists sees if Device exists.
func DeviceExists(index uint8) bool {
	_, err := os.Stat(string(strconv.AppendUint(inputPathSlice, uint64(index-1), 10)))
	return err == nil
}

// Connect sets up a go routine that puts a joysticks events onto registered channels.
// to register channels use the returned HID object's On<xxx>(index) methods.
// Note: only one event, of each type '<xxx>', for each 'index', so re-registering, or deleting, an event stops events going on the old channel.
// It Needs the HID objects ParcelOutEvents() method to be running to perform routing.(so usually in a go routine.)
func Connect(index int) (d *HID, err error) {
	var file *os.File
	if file, err = os.OpenFile(string(strconv.AppendUint(inputPathSlice, uint64(index-1), 10)), os.O_RDONLY, 0); err != nil {
		return
	}
	d = &HID{make(chan osEventRecord), make(map[uint8]button), make(map[uint8]hatAxis), make(map[eventSignature]chan Event)}
	// start thread to read joystick events to the joystick.state osEvent channel
	go eventPipe(file, d.OSEvents)
	d.populate()
	return
}

// fill in the joysticks available events from the synthetic events burst produced initially by the driver.
func (d HID) populate() {
	for buttonNumber, hatNumber, axisNumber := 1, 1, 1; ; {
		evt := <-d.OSEvents
		switch evt.Type {
		case 0x81:
			d.Buttons[evt.Index] = button{uint8(buttonNumber), toDuration(evt.Time), evt.Value != 0}
			buttonNumber++
		case 0x82:
			d.HatAxes[evt.Index] = hatAxis{uint8(hatNumber), uint8(axisNumber), false, toDuration(evt.Time), float32(evt.Value) / maxValue}
			axisNumber++
			if axisNumber > 2 {
				axisNumber = 1
				hatNumber++
			}
		default:
			go func() { d.OSEvents <- evt }() // have to consume a real event to know we reached the end of the synthetic burst, so refire it.
			return
		}
	}
}

// pipe any readable events onto channel.
func eventPipe(r io.Reader, c chan osEventRecord) {
	var evt osEventRecord
	for {
		if binary.Read(r, binary.LittleEndian, &evt) != nil {
			close(c)
			return
		}
		c <- evt
	}
}

func toDuration(m uint32) time.Duration {
	return time.Duration(m) * 1000000
}
