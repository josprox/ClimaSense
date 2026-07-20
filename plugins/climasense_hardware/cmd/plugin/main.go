package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"time"

	"climasense.local/climasense_hardware/internal/bmp280"
	"climasense.local/climasense_hardware/internal/i2c"
)

const protocol = "joss-rpc-v1"

type request struct {
	Protocol string            `json:"protocol"`
	ID       string            `json:"id"`
	Method   string            `json:"method"`
	Args     []json.RawMessage `json:"args"`
}
type rpcError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type response struct {
	ID     string    `json:"id"`
	Result any       `json:"result,omitempty"`
	Error  *rpcError `json:"error,omitempty"`
}

func main() {
	decoder := json.NewDecoder(io.LimitReader(bufio.NewReader(os.Stdin), 1<<20))
	decoder.DisallowUnknownFields()
	var req request
	if err := decoder.Decode(&req); err != nil {
		write(response{Error: &rpcError{Code: "INVALID_REQUEST", Message: err.Error()}})
		return
	}
	if req.Protocol != protocol {
		write(response{ID: req.ID, Error: &rpcError{Code: "INVALID_PROTOCOL", Message: "se esperaba " + protocol}})
		return
	}
	result, err := dispatch(req.Method, req.Args)
	if err != nil {
		write(response{ID: req.ID, Error: &rpcError{Code: errorCode(err), Message: err.Error()}})
		return
	}
	write(response{ID: req.ID, Result: result})
}

func write(value response) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func dispatch(method string, args []json.RawMessage) (any, error) {
	switch method {
	case "ping":
		return map[string]any{"ok": true, "plugin_version": "0.1.0", "platform": runtime.GOOS + "-" + runtime.GOARCH}, nil
	case "i2c_scan":
		return scan(args)
	case "i2c_read_register":
		return readRegister(args)
	case "i2c_write_register":
		return writeRegister(args)
	case "bmp280_read":
		return readBMP280(args)
	case "diagnose":
		return diagnose(args)
	default:
		return nil, fmt.Errorf("metodo desconocido %q", method)
	}
}

func scan(args []json.RawMessage) (any, error) {
	if len(args) != 1 {
		return nil, errors.New("i2c_scan requiere bus")
	}
	bus, err := intArg(args[0], "bus", 0, 255)
	if err != nil {
		return nil, err
	}
	found := []map[string]any{}
	for _, address := range []int{0x76, 0x77} {
		device, openErr := i2c.Open(bus, uint16(address), 500*time.Millisecond)
		if openErr != nil {
			continue
		}
		id := []byte{0}
		readErr := device.ReadRegister(0xD0, id)
		_ = device.Close()
		if readErr == nil {
			found = append(found, map[string]any{"address": fmt.Sprintf("0x%02X", address), "chip_id": fmt.Sprintf("0x%02X", id[0]), "sensor_type": sensorType(id[0])})
		}
	}
	return found, nil
}

func readRegister(args []json.RawMessage) (any, error) {
	if len(args) != 5 {
		return nil, errors.New("i2c_read_register requiere bus, address, register, length, timeout_ms")
	}
	bus, address, register, length, timeout, err := transactionArgs(args)
	if err != nil {
		return nil, err
	}
	device, err := i2c.Open(bus, uint16(address), time.Duration(timeout)*time.Millisecond)
	if err != nil {
		return nil, err
	}
	defer device.Close()
	data := make([]byte, length)
	if err := device.ReadRegister(byte(register), data); err != nil {
		return nil, err
	}
	values := make([]int, len(data))
	for i, value := range data {
		values[i] = int(value)
	}
	return values, nil
}

func writeRegister(args []json.RawMessage) (any, error) {
	if len(args) != 5 {
		return nil, errors.New("i2c_write_register requiere bus, address, register, data, timeout_ms")
	}
	bus, err := intArg(args[0], "bus", 0, 255)
	if err != nil {
		return nil, err
	}
	address, err := intArg(args[1], "address", 0x03, 0x77)
	if err != nil {
		return nil, err
	}
	register, err := intArg(args[2], "register", 0, 255)
	if err != nil {
		return nil, err
	}
	var raw []int
	if err := json.Unmarshal(args[3], &raw); err != nil || len(raw) > 4096 {
		return nil, errors.New("data debe ser un array de hasta 4096 bytes")
	}
	data := make([]byte, len(raw))
	for index, value := range raw {
		if value < 0 || value > 255 {
			return nil, fmt.Errorf("data[%d] fuera de rango", index)
		}
		data[index] = byte(value)
	}
	timeout, err := intArg(args[4], "timeout_ms", 10, 60000)
	if err != nil {
		return nil, err
	}
	device, err := i2c.Open(bus, uint16(address), time.Duration(timeout)*time.Millisecond)
	if err != nil {
		return nil, err
	}
	defer device.Close()
	if err := device.WriteRegister(byte(register), data); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "bytes_written": len(data)}, nil
}

func readBMP280(args []json.RawMessage) (any, error) {
	if len(args) != 3 {
		return nil, errors.New("bmp280_read requiere bus, address y config")
	}
	bus, err := intArg(args[0], "bus", 0, 255)
	if err != nil {
		return nil, err
	}
	addresses, err := addressesArg(args[1])
	if err != nil {
		return nil, err
	}
	config, err := configArg(args[2])
	if err != nil {
		return nil, err
	}
	var failures []string
	for _, address := range addresses {
		device, openErr := i2c.Open(bus, uint16(address), config.Timeout)
		if openErr != nil {
			failures = append(failures, openErr.Error())
			continue
		}
		sensor, sensorErr := bmp280.New(device, uint16(address), config)
		if sensorErr != nil {
			_ = device.Close()
			failures = append(failures, sensorErr.Error())
			continue
		}
		measurement, measureErr := sensor.Measure()
		closeErr := sensor.Close()
		if measureErr != nil {
			failures = append(failures, measureErr.Error())
			continue
		}
		if closeErr != nil {
			return nil, closeErr
		}
		return measurement, nil
	}
	return nil, fmt.Errorf("BMP280 no disponible: %v", failures)
}

func diagnose(args []json.RawMessage) (any, error) {
	if len(args) != 1 {
		return nil, errors.New("diagnose requiere bus")
	}
	bus, err := intArg(args[0], "bus", 0, 255)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/dev/i2c-%d", bus)
	stat, statErr := os.Stat(path)
	result := map[string]any{"platform": runtime.GOOS + "-" + runtime.GOARCH, "bus": bus, "path": path, "exists": statErr == nil}
	if statErr != nil {
		result["error"] = statErr.Error()
		return result, nil
	}
	result["mode"] = stat.Mode().String()
	devices, _ := scan([]json.RawMessage{json.RawMessage(strconv.Itoa(bus))})
	result["devices"] = devices
	return result, nil
}

func transactionArgs(args []json.RawMessage) (int, int, int, int, int, error) {
	bus, err := intArg(args[0], "bus", 0, 255)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	address, err := intArg(args[1], "address", 0x03, 0x77)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	register, err := intArg(args[2], "register", 0, 255)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	length, err := intArg(args[3], "length", 1, 4096)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	timeout, err := intArg(args[4], "timeout_ms", 10, 60000)
	return bus, address, register, length, timeout, err
}

func intArg(raw json.RawMessage, name string, min, max int) (int, error) {
	var value int
	if err := json.Unmarshal(raw, &value); err != nil || value < min || value > max {
		return 0, fmt.Errorf("%s debe ser entero entre %d y %d", name, min, max)
	}
	return value, nil
}

func addressesArg(raw json.RawMessage) ([]int, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	if text, ok := value.(string); ok {
		if text == "auto" {
			return []int{0x76, 0x77}, nil
		}
		parsed, err := strconv.ParseInt(text, 0, 16)
		if err != nil {
			return nil, errors.New("address debe ser auto, 0x76 o 0x77")
		}
		value = float64(parsed)
	}
	number, ok := value.(float64)
	if !ok || (int(number) != 0x76 && int(number) != 0x77) {
		return nil, errors.New("address debe ser auto, 0x76 o 0x77")
	}
	return []int{int(number)}, nil
}

func configArg(raw json.RawMessage) (bmp280.Config, error) {
	cfg := bmp280.DefaultConfig()
	var values map[string]int
	if err := json.Unmarshal(raw, &values); err != nil {
		return cfg, errors.New("config debe ser un mapa")
	}
	if v, ok := values["temperature_oversampling"]; ok {
		cfg.TemperatureOversampling = v
	}
	if v, ok := values["pressure_oversampling"]; ok {
		cfg.PressureOversampling = v
	}
	if v, ok := values["filter"]; ok {
		cfg.Filter = v
	}
	if v, ok := values["timeout_ms"]; ok {
		if v < 10 || v > 60000 {
			return cfg, errors.New("timeout_ms fuera de rango")
		}
		cfg.Timeout = time.Duration(v) * time.Millisecond
	}
	return cfg, nil
}

func sensorType(id byte) string {
	if id == bmp280.ChipIDBMP280 {
		return "bmp280"
	}
	if id == bmp280.ChipIDBME280 {
		return "bme280"
	}
	return "unknown"
}
func errorCode(err error) string {
	message := err.Error()
	if len(message) >= 7 && message[:7] == "metodo " {
		return "METHOD_NOT_FOUND"
	}
	return "HARDWARE_ERROR"
}
