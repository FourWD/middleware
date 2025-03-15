package common

import "github.com/tarm/serial"

func IsSerialPortConnect(serialPort *serial.Port) bool {
	if serialPort == nil {
		return false
	}

	_, err := serialPort.Write([]byte{0x00}) // Send a dummy byte
	return err == nil
}
