package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/holoplot/go-evdev"
	"go.bug.st/serial"
)

type PoseCommand struct {
	notifyC chan struct{}
	mu      sync.Mutex
	velPon  int8
	velSide int8
	velTilt int8
}

func pollKeyPresses(cmd *PoseCommand) {
	d, err := evdev.Open("/dev/input/event3")
	if err != nil {
		fmt.Printf("Cannot read %s: %v\n", os.Args[1], err)
		return
	}

	for {
		e, err := d.ReadOne()
		if err != nil {
			fmt.Printf("Error reading from device: %v\n", err)
			return
		}

		// Useful for debugging
		// fmt.Printf("Event Code %d, Value %d\n", e.Code, e.Value)

		switch e.Type {
		case evdev.EV_SYN:
			switch e.Code {
			case evdev.SYN_MT_REPORT:
				// Cluster of frames have been processed
			case evdev.SYN_DROPPED:
			default:
			}
		default:

			const CodeA = 30
			const CodeD = 32
			const CodeW = 17
			const CodeS = 31
			const CodeDownArrow = 108
			const CodeUpArrow = 103
			switch e.Code {
			case CodeA, CodeD, CodeW, CodeS, CodeDownArrow, CodeUpArrow:
				// Ensure it's a relevant key
			default:
				continue
			}

			const ValueKeyOn = 1
			const ValueKeyOff = 0

			switch e.Value {
				case ValueKeyOn, ValueKeyOff:
					// Ensure it's either on/off
				default:
					continue
			}
			
			cmd.mu.Lock()

			switch e.Code{
			case CodeW:
				if e.Value == ValueKeyOn{
					cmd.velTilt = 127
				} else{
					cmd.velTilt = 0
				}
			case CodeS:
				if e.Value == ValueKeyOn{
					cmd.velTilt = -127
				} else{
					cmd.velTilt = 0
				}
			case CodeA:
				if e.Value == ValueKeyOn{
					cmd.velSide = 127
				} else{
					cmd.velSide = 0
				}
			case CodeD:
				if e.Value == ValueKeyOn{
					cmd.velSide = -127
				} else{
					cmd.velSide = 0
				}
			case CodeDownArrow:
				if e.Value == ValueKeyOn{
					cmd.velPon = -127
				} else{
					cmd.velPon = 0
				}
			case CodeUpArrow:
				if e.Value == ValueKeyOn{
					cmd.velPon = 127
				} else{
					cmd.velPon = 0
				}
				
			}

			// fmt.Println("Key press")
			select {
			case cmd.notifyC <- struct{}{}:
			default:
			}

			cmd.mu.Unlock()

		}
	}
}

func handleSerial(cmd *PoseCommand) {
	mode := &serial.Mode{BaudRate: 115200}
	port, err := serial.Open("/dev/ttyACM0", mode)
	if err != nil {
		fmt.Printf("Cannot read %s: %v\n", os.Args[1], err)
		return
	}
	defer port.Close()

	for {
		<-cmd.notifyC
		// fmt.Println("Received notification")
		cmd.mu.Lock()
		velPon := cmd.velPon
		velSide := cmd.velSide
		velTilt := cmd.velTilt
		cmd.mu.Unlock()

		payload := fmt.Sprintf("%d,%d,%d\n", velPon, velSide, velTilt)
		fmt.Printf("pon:%d, side:%d, tilt:%d\n", velPon, velSide, velTilt)
		_, err := port.Write([]byte(payload))
		if err != nil {
			fmt.Printf("Cannot write to serial port: %v\n", err)
			return
		}

		go func(){
			buff := make([]byte, 100)
			for {
				n, err := port.Read(buff)
				if err != nil {
					break
				}
				if n == 0 {
					fmt.Println("\nEOF")
					break
				}
					fmt.Printf("Teensy 4.1: %v", string(buff[:n]))
			}
		}()

		// n, err := port.Read(buff)
		// if err != nil {
		// 	// log.Fatal(err)
		// 	break
		// }
		// if n == 0 {
		// 	fmt.Println("\nEOF")
		// 	break
		// }
		// fmt.Printf("RECEIVED: %v", string(buff[:n]))
		//
		// n, err := port.Write([]byte("10,20,30\n\r"))
		// if err != nil {
		// 	loaddaswsssswasddadwsadwswadswswdaddsswswadwsadadasg.Fatal(err)
		// }wsssswadws
		// fmt.Printf("Sent %v bytes\n", n)swadwsw
		//
		// buff := make([]byte, 100)

	}
}

func main() {
	cmd := &PoseCommand{
		notifyC: make(chan struct{}, 1),
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		pollKeyPresses(cmd)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		handleSerial(cmd)
	}()

	wg.Wait()
}
