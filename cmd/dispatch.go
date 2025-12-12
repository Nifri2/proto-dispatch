package cmd

import (
	"fmt"
	"machine"
	"math/rand"
	"time"
)

const (
	ProjectedFPS = 47
)

// Reworked Protocol to support 2 animation channels (eye + mouth)
// Example: [Header(0xAA), address, command, animID_eye, animID_mouth, checksum]

func RunDispatcher(config Settings, uart *machine.UART, led machine.Pin) {
	fmt.Println("Starting Dispatcher Loop")

	// Channel to serialize UART writes
	// stores [6]byte{header, address, command, animID_eye, animID_mouth, checksum}
	uartChan := make(chan [6]byte, 10)

	// Goroutine to handle UART writes
	go func() {
		for packet := range uartChan {
			// Small delay between writes to ensure receiver can keep up if needed                                                                                                                                                                                                              │
			// time.Sleep(5 * time.Millisecond)
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

	for _, w := range workers {
		go func(workerAddr Address) {
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerAddr)))

			// Init to Idle
			sendPacket(workerAddr, Cmd_DisplayAnim, Anim_EyeIdle, Anim_MouthIdle)

			for {
				// Random sleep 2-6 seconds
				sleepDuration := time.Duration(2000+r.Intn(4001)) * time.Millisecond
				time.Sleep(sleepDuration)

				// Send Blink
				// fmt.Printf("Blinking Worker %d\n", workerAddr)
				sendPacket(workerAddr, Cmd_DisplayAnim, Anim_EyeBlink, Anim_MouthIdle)

				// Blink duration = 50 frames at ProjectedFPS

				blinkDuration := time.Duration(50*1000/ProjectedFPS) * time.Millisecond
				time.Sleep(blinkDuration)
				// time.Sleep(200 * time.Millisecond)

				// Send Idle
				sendPacket(workerAddr, Cmd_DisplayAnim, Anim_EyeIdle, Anim_MouthIdle)
			}
		}(w)

	}

	// Block forever
	select {}
}
