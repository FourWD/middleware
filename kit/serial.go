package kit

import "github.com/tarm/serial"

func IsSerialPortConnected(port *serial.Port) bool {
	if port == nil {
		return false
	}

	_, err := port.Write([]byte{0x00})
	return err == nil
}
