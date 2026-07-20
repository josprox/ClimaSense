//go:build linux

package i2c

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	i2cRetries = 0x0701
	i2cTimeout = 0x0702
	i2cSlave   = 0x0703
)

type linuxDevice struct {
	mu      sync.Mutex
	file    *os.File
	path    string
	address uint16
	closed  bool
}

func ioctl(fd uintptr, request uintptr, value uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, request, value)
	if errno != 0 {
		return errno
	}
	return nil
}

func openPlatform(bus int, address uint16, timeout time.Duration) (Device, error) {
	path := fmt.Sprintf("/dev/i2c-%d", bus)
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("abrir %s: %w", path, err)
	}
	closeWith := func(operation string, err error) (Device, error) {
		_ = file.Close()
		return nil, fmt.Errorf("%s en %s: %w", operation, path, err)
	}
	if err := ioctl(file.Fd(), i2cSlave, uintptr(address)); err != nil {
		return closeWith(fmt.Sprintf("seleccionar direccion 0x%02X", address), err)
	}
	if timeout > 0 {
		units := uintptr((timeout + 9*time.Millisecond) / (10 * time.Millisecond))
		if units < 1 {
			units = 1
		}
		if err := ioctl(file.Fd(), i2cTimeout, units); err != nil {
			return closeWith("configurar timeout", err)
		}
	}
	if err := ioctl(file.Fd(), i2cRetries, uintptr(2)); err != nil {
		return closeWith("configurar reintentos", err)
	}
	return &linuxDevice{file: file, path: path, address: address}, nil
}

func (d *linuxDevice) ReadRegister(register byte, destination []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return ErrClosed
	}
	if len(destination) == 0 {
		return errorsNew("la lectura requiere al menos un byte")
	}
	if n, err := d.file.Write([]byte{register}); err != nil || n != 1 {
		return fmt.Errorf("seleccionar registro 0x%02X: bytes=%d: %w", register, n, err)
	}
	if n, err := d.file.Read(destination); err != nil || n != len(destination) {
		return fmt.Errorf("leer registro 0x%02X: bytes=%d/%d: %w", register, n, len(destination), err)
	}
	return nil
}

func (d *linuxDevice) WriteRegister(register byte, data []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return ErrClosed
	}
	payload := make([]byte, len(data)+1)
	payload[0] = register
	copy(payload[1:], data)
	if n, err := d.file.Write(payload); err != nil || n != len(payload) {
		return fmt.Errorf("escribir registro 0x%02X: bytes=%d/%d: %w", register, n, len(payload), err)
	}
	return nil
}

func (d *linuxDevice) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil
	}
	d.closed = true
	return d.file.Close()
}

// Keep unsafe referenced in the Linux build where syscall.Syscall uses uintptr
// arguments and static analyzers otherwise lose the ioctl pointer context.
var _ unsafe.Pointer

func errorsNew(message string) error { return fmt.Errorf("%s", message) }
