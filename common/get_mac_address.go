package common

import (
	"errors"
	"net"
)

func GetMacAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.HardwareAddr != nil {
			return iface.HardwareAddr.String(), nil
		}
	}

	return "", errors.New("no MAC address found")
}
