package bmp280

import (
	"encoding/binary"
	"errors"
	"math"
	"sync"
	"testing"
)

type fakeDevice struct {
	mu        sync.Mutex
	registers [256]byte
	closed    bool
	readErr   error
}

func (d *fakeDevice) ReadRegister(register byte, destination []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.readErr != nil {
		return d.readErr
	}
	copy(destination, d.registers[int(register):int(register)+len(destination)])
	return nil
}

func (d *fakeDevice) WriteRegister(register byte, data []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	copy(d.registers[int(register):], data)
	return nil
}

func (d *fakeDevice) Close() error { d.closed = true; return nil }

func calibratedDevice(chipID byte) *fakeDevice {
	d := &fakeDevice{}
	d.registers[regID] = chipID
	c := Calibration{T1: 27504, T2: 26435, T3: -1000, P1: 36477, P2: -10685, P3: 3024, P4: 2855, P5: 140, P6: -7, P7: 15500, P8: -14600, P9: 6000}
	values := []uint16{c.T1, uint16(c.T2), uint16(c.T3), c.P1, uint16(c.P2), uint16(c.P3), uint16(c.P4), uint16(c.P5), uint16(c.P6), uint16(c.P7), uint16(c.P8), uint16(c.P9)}
	for index, value := range values {
		binary.LittleEndian.PutUint16(d.registers[regCalib+byte(index*2):], value)
	}
	return d
}

func setRaw(d *fakeDevice, pressure, temperature int32) {
	d.registers[regPressure] = byte(pressure >> 12)
	d.registers[regPressure+1] = byte(pressure >> 4)
	d.registers[regPressure+2] = byte(pressure << 4)
	d.registers[regPressure+3] = byte(temperature >> 12)
	d.registers[regPressure+4] = byte(temperature >> 4)
	d.registers[regPressure+5] = byte(temperature << 4)
}

func TestDatasheetCompensation(t *testing.T) {
	d := calibratedDevice(ChipIDBMP280)
	setRaw(d, 415148, 519888)
	sensor, err := New(d, 0x76, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	measurement, err := sensor.Measure()
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(measurement.TemperatureC-25.08) > 0.01 {
		t.Fatalf("temperature = %.2f", measurement.TemperatureC)
	}
	if math.Abs(measurement.PressurePa-100653.27) > 0.1 {
		t.Fatalf("pressure = %.2f", measurement.PressurePa)
	}
	if measurement.Address != "0x76" || measurement.Status != "ok" {
		t.Fatalf("measurement = %#v", measurement)
	}
}

func TestRejectsBME280AndUnknownChip(t *testing.T) {
	if _, err := New(calibratedDevice(ChipIDBME280), 0x76, DefaultConfig()); err == nil {
		t.Fatal("BME280 was accepted")
	}
	if _, err := New(calibratedDevice(0x00), 0x76, DefaultConfig()); err == nil {
		t.Fatal("unknown chip was accepted")
	}
}

func TestPropagatesBusFailure(t *testing.T) {
	d := calibratedDevice(ChipIDBMP280)
	d.readErr = errors.New("remote I/O error")
	if _, err := New(d, 0x76, DefaultConfig()); err == nil {
		t.Fatal("bus failure was ignored")
	}
}

func TestCloseDelegatesToI2CDevice(t *testing.T) {
	d := calibratedDevice(ChipIDBMP280)
	sensor, err := New(d, 0x77, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if err := sensor.Close(); err != nil {
		t.Fatal(err)
	}
	if !d.closed {
		t.Fatal("device was not closed")
	}
}
