package bmp280

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"

	"climasense.local/climasense_hardware/internal/i2c"
)

const (
	ChipIDBMP280 = 0x58
	ChipIDBME280 = 0x60
	regID        = 0xD0
	regReset     = 0xE0
	regStatus    = 0xF3
	regControl   = 0xF4
	regConfig    = 0xF5
	regPressure  = 0xF7
	regCalib     = 0x88
)

type Config struct {
	TemperatureOversampling int
	PressureOversampling    int
	Filter                  int
	Timeout                 time.Duration
}

func DefaultConfig() Config {
	return Config{TemperatureOversampling: 1, PressureOversampling: 4, Filter: 2, Timeout: time.Second}
}

type Calibration struct {
	T1                             uint16
	T2, T3                         int16
	P1                             uint16
	P2, P3, P4, P5, P6, P7, P8, P9 int16
}

type Measurement struct {
	TemperatureC float64 `json:"temperature_c"`
	PressurePa   float64 `json:"pressure_pa"`
	PressureHPa  float64 `json:"pressure_hpa"`
	Address      string  `json:"sensor_address"`
	Status       string  `json:"sensor_status"`
	MeasuredAt   string  `json:"measured_at"`
}

type Sensor struct {
	device      i2c.Device
	address     uint16
	config      Config
	calibration Calibration
}

func New(device i2c.Device, address uint16, config Config) (*Sensor, error) {
	if device == nil {
		return nil, errors.New("dispositivo I2C nulo")
	}
	if _, err := oversamplingBits(config.TemperatureOversampling); err != nil {
		return nil, fmt.Errorf("sobremuestreo de temperatura: %w", err)
	}
	if _, err := oversamplingBits(config.PressureOversampling); err != nil {
		return nil, fmt.Errorf("sobremuestreo de presion: %w", err)
	}
	filter, err := filterBits(config.Filter)
	if err != nil {
		return nil, err
	}
	id := []byte{0}
	if err := device.ReadRegister(regID, id); err != nil {
		return nil, fmt.Errorf("leer chip ID: %w", err)
	}
	switch id[0] {
	case ChipIDBMP280:
	case ChipIDBME280:
		return nil, errors.New("se detecto BME280 (chip ID 0x60), no BMP280")
	default:
		return nil, fmt.Errorf("chip ID inesperado 0x%02X", id[0])
	}
	if err := device.WriteRegister(regReset, []byte{0xB6}); err != nil {
		return nil, fmt.Errorf("reiniciar BMP280: %w", err)
	}
	time.Sleep(3 * time.Millisecond)
	calibrationData := make([]byte, 24)
	if err := device.ReadRegister(regCalib, calibrationData); err != nil {
		return nil, fmt.Errorf("leer calibracion BMP280: %w", err)
	}
	calibration := decodeCalibration(calibrationData)
	if calibration.T1 == 0 || calibration.P1 == 0 {
		return nil, errors.New("calibracion BMP280 invalida")
	}
	if err := device.WriteRegister(regConfig, []byte{filter << 2}); err != nil {
		return nil, fmt.Errorf("configurar filtro BMP280: %w", err)
	}
	if config.Timeout <= 0 {
		config.Timeout = time.Second
	}
	return &Sensor{device: device, address: address, config: config, calibration: calibration}, nil
}

func (s *Sensor) Measure() (Measurement, error) {
	tBits, _ := oversamplingBits(s.config.TemperatureOversampling)
	pBits, _ := oversamplingBits(s.config.PressureOversampling)
	control := tBits<<5 | pBits<<2 | 0x01
	if err := s.device.WriteRegister(regControl, []byte{control}); err != nil {
		return Measurement{}, fmt.Errorf("iniciar medicion forzada: %w", err)
	}
	deadline := time.Now().Add(s.config.Timeout)
	for {
		status := []byte{0}
		if err := s.device.ReadRegister(regStatus, status); err != nil {
			return Measurement{}, fmt.Errorf("consultar estado BMP280: %w", err)
		}
		if status[0]&0x08 == 0 {
			break
		}
		if time.Now().After(deadline) {
			return Measurement{}, errors.New("timeout esperando medicion BMP280")
		}
		time.Sleep(2 * time.Millisecond)
	}
	raw := make([]byte, 6)
	if err := s.device.ReadRegister(regPressure, raw); err != nil {
		return Measurement{}, fmt.Errorf("leer medicion BMP280: %w", err)
	}
	rawPressure := int32(raw[0])<<12 | int32(raw[1])<<4 | int32(raw[2])>>4
	rawTemperature := int32(raw[3])<<12 | int32(raw[4])<<4 | int32(raw[5])>>4
	temperature, fine := compensateTemperature(rawTemperature, s.calibration)
	pressure, err := compensatePressure(rawPressure, fine, s.calibration)
	if err != nil {
		return Measurement{}, err
	}
	if temperature < -40 || temperature > 85 || pressure < 30000 || pressure > 110000 {
		return Measurement{}, fmt.Errorf("medicion fuera de rango: %.2f C, %.2f Pa", temperature, pressure)
	}
	return Measurement{
		TemperatureC: round(temperature, 2), PressurePa: round(pressure, 2), PressureHPa: round(pressure/100, 2),
		Address: fmt.Sprintf("0x%02X", s.address), Status: "ok", MeasuredAt: time.Now().UTC().Format(time.RFC3339Nano),
	}, nil
}

func (s *Sensor) Close() error { return s.device.Close() }

func decodeCalibration(data []byte) Calibration {
	u := func(offset int) uint16 { return binary.LittleEndian.Uint16(data[offset : offset+2]) }
	i := func(offset int) int16 { return int16(u(offset)) }
	return Calibration{T1: u(0), T2: i(2), T3: i(4), P1: u(6), P2: i(8), P3: i(10), P4: i(12), P5: i(14), P6: i(16), P7: i(18), P8: i(20), P9: i(22)}
}

func compensateTemperature(raw int32, c Calibration) (float64, float64) {
	var1 := (float64(raw)/16384.0 - float64(c.T1)/1024.0) * float64(c.T2)
	var2 := (float64(raw)/131072.0 - float64(c.T1)/8192.0)
	var2 = var2 * var2 * float64(c.T3)
	fine := var1 + var2
	return fine / 5120.0, fine
}

func compensatePressure(raw int32, fine float64, c Calibration) (float64, error) {
	var1 := fine/2.0 - 64000.0
	var2 := var1 * var1 * float64(c.P6) / 32768.0
	var2 += var1 * float64(c.P5) * 2.0
	var2 = var2/4.0 + float64(c.P4)*65536.0
	var1 = (float64(c.P3)*var1*var1/524288.0 + float64(c.P2)*var1) / 524288.0
	var1 = (1.0 + var1/32768.0) * float64(c.P1)
	if math.Abs(var1) < 0.000001 {
		return 0, errors.New("calibracion BMP280 invalida: divisor de presion cero")
	}
	pressure := 1048576.0 - float64(raw)
	pressure = (pressure - var2/4096.0) * 6250.0 / var1
	var1 = float64(c.P9) * pressure * pressure / 2147483648.0
	var2 = pressure * float64(c.P8) / 32768.0
	return pressure + (var1+var2+float64(c.P7))/16.0, nil
}

func oversamplingBits(value int) (byte, error) {
	switch value {
	case 0:
		return 0, nil
	case 1:
		return 1, nil
	case 2:
		return 2, nil
	case 4:
		return 3, nil
	case 8:
		return 4, nil
	case 16:
		return 5, nil
	}
	return 0, fmt.Errorf("valor %d; permitidos: 0, 1, 2, 4, 8, 16", value)
}

func filterBits(value int) (byte, error) {
	switch value {
	case 0:
		return 0, nil
	case 2:
		return 1, nil
	case 4:
		return 2, nil
	case 8:
		return 3, nil
	case 16:
		return 4, nil
	}
	return 0, fmt.Errorf("filtro invalido %d; permitidos: 0, 2, 4, 8, 16", value)
}

func round(value float64, places int) float64 {
	scale := math.Pow10(places)
	return math.Round(value*scale) / scale
}
