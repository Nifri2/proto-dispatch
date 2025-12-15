package cmd

import (
	"fmt"
	"machine"
	"math/rand"
	"time"
)

const (
	ProjectedFPS = 47
	Radio_Pin_0  = machine.GP19
	Radio_Pin_1  = machine.GP18
	Radio_Pin_2  = machine.GP17
	Radio_Pin_3  = machine.GP16
)

var Radio_Pins = []machine.Pin{Radio_Pin_0, Radio_Pin_1, Radio_Pin_2, Radio_Pin_3}

var Radio_Map = map[string]int{
	// Single Press (N_0)
	"Radio_A_0": 0x00,
	"Radio_B_0": 0x01,
	"Radio_C_0": 0x02,
	"Radio_D_0": 0x03,

	// Double Press (N_N) - Starting at 0x04
	// P1 = A
	"Radio_A_A": 0x04,
	"Radio_A_B": 0x05,
	"Radio_A_C": 0x06,
	"Radio_A_D": 0x07,

	// P1 = B
	"Radio_B_A": 0x08,
	"Radio_B_B": 0x09,
	"Radio_B_C": 0x0A,
	"Radio_B_D": 0x0B,

	// P1 = C
	"Radio_C_A": 0x0C,
	"Radio_C_B": 0x0D,
	"Radio_C_C": 0x0E,
	"Radio_C_D": 0x0F,

	// P1 = D
	"Radio_D_A": 0x10,
	"Radio_D_B": 0x11,
	"Radio_D_C": 0x12,
	"Radio_D_D": 0x13,
}

// Reworked Protocol to support 2 animation channels (eye + mouth)
// Example: [Header(0xAA), address, command, animID_eye, animID_mouth, checksum]

func RunDispatcher(config Settings, uart *machine.UART, led machine.Pin) {
	fmt.Println("Starting Dispatcher Loop")

	// Configure Watchdog (5s timeout)
	machine.Watchdog.Configure(machine.WatchdogConfig{TimeoutMillis: 5000})

	// Channel to serialize UART writes
	// stores [6]byte{header, address, command, animID_eye, animID_mouth, checksum}
	uartChan := make(chan [6]byte, 10)
	radioChan := make(chan []byte, 10)

	// Radio handling goroutine
	// We have 4 pins. To get 16 selections, we detect a sequence of presses.
	// 1. Detect First Press (P1)
	// 2. Wait up to 1 second for Second Press (P2)
	// 3. If P2 occurs: Value = (P1 << 2) | P2. (Range 0-15)
	// 4. If Timeout: Value = P1. (Range 0-3)
	go func() {
		for _, pin := range Radio_Pins {
			pin.Configure(machine.PinConfig{Mode: machine.PinInput})
		}

		getPressedPin := func() int {
			for i, pin := range Radio_Pins {
				if pin.Get() {
					return i
				}
			}
			return -1
		}

		for {
			// 1. Wait for first press
			var p1 int = -1
			for {
				p1 = getPressedPin()
				if p1 != -1 {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}

			// Wait for release (Debounce / Separation)
			for getPressedPin() != -1 {
				time.Sleep(10 * time.Millisecond)
			}

			// 2. Wait up to 1 second for second press
			var p2 int = -1
			timeout := time.Now().Add(1 * time.Second)
			for time.Now().Before(timeout) {
				p := getPressedPin()
				if p != -1 {
					p2 = p
					break
				}
				time.Sleep(10 * time.Millisecond)
			}

			var result byte
			if p2 != -1 {
				// Double press detected
				// Wait for release of second button
				for getPressedPin() != -1 {
					time.Sleep(10 * time.Millisecond)
				}
				// Map A+A (0,0) to 4
				// Value = 4 + (p1 * 4) + p2
				result = byte(4 + (p1 << 2) | p2)
			} else {
				// Single press (Timeout)
				// A=0, B=1, C=2, D=3
				result = byte(p1)
			}

			// Send to channel
			radioChan <- []byte{result}
		}
	}()

	// Output Radio handling goroutine
	go func() {
		for packet := range radioChan {
			// debug print
			if packet != nil && len(packet) > 0 {
				fmt.Printf("Radio Packet Detected: 0x%02X\n", packet[0])
			}
		}
	}()

	// Goroutine to handle UART writes
	go func() {
		for packet := range uartChan {
			uart.Write(packet[:])
		}
	}()

	sendPacket := func(addr Address, cmd Command, eye, mouth AnimationID) {
		header := byte(0xAA)
		a := byte(addr)
		c := byte(cmd)
		e := byte(eye)
		m := byte(mouth)
		checksum := a + c + e + m
		uartChan <- [6]byte{header, a, c, e, m, checksum}
	}

	// Workers to control
	workers := []Address{Worker_0, Worker_1, Worker_2, Worker_3}

	// Initialize random number generator once
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	inter := 0 // Counter for 'still alive' messages

	// Main Dispatcher Loop (user's desired logic)
	for {
		machine.Watchdog.Update() // Feed watchdog

		// Random sleep 2-6 seconds
		sleepDuration := time.Duration(2000+r.Intn(4001)) * time.Millisecond
		time.Sleep(sleepDuration)

		for _, workerAddr := range workers {
			sendPacket(workerAddr, Cmd_DisplayAnim, Anim_EyeBlink, Anim_MouthIdle)
		}

		// Blink duration = 50 frames at ProjectedFPS
		blinkDuration := time.Duration(50*1000/ProjectedFPS) * time.Millisecond
		time.Sleep(blinkDuration)

		// Send Idle
		for _, workerAddr := range workers {
			sendPacket(workerAddr, Cmd_DisplayAnim, Anim_EyeIdle, Anim_MouthIdle)
		}

		inter++
		fmt.Println("still alive", inter)
	}
}
