package i2c

import (
	"errors"
	"fmt"
	"time"
)

var ErrClosed = errors.New("dispositivo I2C cerrado")

type Device interface {
	ReadRegister(register byte, destination []byte) error
	WriteRegister(register byte, data []byte) error
	Close() error
}

func Open(bus int, address uint16, timeout time.Duration) (Device, error) {
	if bus < 0 {
		return nil, fmt.Errorf("bus I2C invalido: %d", bus)
	}
	if address < 0x03 || address > 0x77 {
		return nil, fmt.Errorf("direccion I2C fuera del rango de 7 bits: 0x%X", address)
	}
	return openPlatform(bus, address, timeout)
}
