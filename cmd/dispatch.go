package cmd

import (
	"fmt"
	"machine"
	"time"
)

func RunDispatcher(config Settings, uart *machine.UART, led machine.Pin) {
	// Send commands to Workers
	// Worker 1 LED On wait for 1 second LED Off repeat for Worker 2, back to Worker 1
	fmt.Println("Starting Dispatcher Loop")

	delay := time.Millisecond * 10

	for {
		// Worker 1 LED ON
		uart.Write([]byte{byte(Worker_1), byte(Cmd_LedOn)})
		time.Sleep(delay)
		led.High()

		// Worker 1 LED OFF
		uart.Write([]byte{byte(Worker_1), byte(Cmd_LedOff)})
		time.Sleep(delay)
		led.Low()

		// Worker 2 LED ON
		uart.Write([]byte{byte(Worker_2), byte(Cmd_LedOn)})
		time.Sleep(delay)
		led.High()

		// Worker 2 LED OFF
		uart.Write([]byte{byte(Worker_2), byte(Cmd_LedOff)})
		time.Sleep(delay)
		led.Low()
	}
}
