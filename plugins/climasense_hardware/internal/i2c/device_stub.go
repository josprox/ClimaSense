//go:build !linux

package i2c

import (
	"fmt"
	"runtime"
	"time"
)

func openPlatform(bus int, address uint16, timeout time.Duration) (Device, error) {
	return nil, fmt.Errorf("I2C real requiere Linux; plataforma actual: %s", runtime.GOOS)
}
