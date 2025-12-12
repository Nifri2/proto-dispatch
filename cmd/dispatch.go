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

	// Configure Watchdog (5s timeout)
	machine.Watchdog.Configure(machine.WatchdogConfig{TimeoutMillis: 5000})

	// Channel to serialize UART writes
	// stores [6]byte{header, address, command, animID_eye, animID_mouth, checksum}
	uartChan := make(chan [6]byte, 10)

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
