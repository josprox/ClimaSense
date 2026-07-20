package main

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"climasense.local/climasense_transport/internal/protocol"
)

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
	decoder := json.NewDecoder(io.LimitReader(os.Stdin, 1<<20))
	decoder.DisallowUnknownFields()
	var req request
	if err := decoder.Decode(&req); err != nil {
		write(response{Error: &rpcError{Code: "INVALID_REQUEST", Message: err.Error()}})
		return
	}
	if req.Protocol != "joss-rpc-v1" {
		write(response{ID: req.ID, Error: &rpcError{Code: "INVALID_PROTOCOL", Message: "se esperaba joss-rpc-v1"}})
		return
	}
	result, err := dispatch(req.Method, req.Args)
	if err != nil {
		write(response{ID: req.ID, Error: &rpcError{Code: "PROTOCOL_ERROR", Message: err.Error()}})
		return
	}
	write(response{ID: req.ID, Result: result})
}
func write(value response) {
	if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func dispatch(method string, args []json.RawMessage) (any, error) {
	switch method {
	case "ping":
		return map[string]any{"ok": true, "plugin_version": "0.2.0", "platform": runtime.GOOS + "-" + runtime.GOARCH}, nil
	case "hash_token":
		if len(args) != 1 {
			return nil, errors.New("hash_token requiere token")
		}
		var token string
		if json.Unmarshal(args[0], &token) != nil {
			return nil, errors.New("token invalido")
		}
		return protocol.HashToken(token), nil
	case "secure_equal":
		if len(args) != 2 {
			return nil, errors.New("secure_equal requiere dos strings")
		}
		var left, right string
		if json.Unmarshal(args[0], &left) != nil || json.Unmarshal(args[1], &right) != nil {
			return nil, errors.New("secure_equal requiere strings")
		}
		return len(left) == len(right) && subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1, nil
	case "random_token":
		if len(args) != 1 {
			return nil, errors.New("random_token requiere bytes")
		}
		var size int
		if json.Unmarshal(args[0], &size) != nil {
			return nil, errors.New("bytes invalido")
		}
		return protocol.RandomToken(size)
	case "sign":
		return sign(args)
	case "verify":
		return verify(args)
	case "verify_hash":
		return verifyHash(args)
	case "send":
		return send(args)
	case "ws_send":
		return wsSend(args)
	case "activate":
		return activate(args)
	case "weather_current":
		return weatherCurrent(args)
	default:
		return nil, fmt.Errorf("metodo desconocido %q", method)
	}
}

func wsSend(args []json.RawMessage) (any, error) {
	if len(args) != 7 {
		return nil, errors.New("ws_send requiere 7 argumentos")
	}
	var base, path, device, token string
	var sequence int64
	var payload any
	var timeout int
	if json.Unmarshal(args[0], &base) != nil || json.Unmarshal(args[1], &path) != nil || json.Unmarshal(args[2], &device) != nil || json.Unmarshal(args[3], &token) != nil || json.Unmarshal(args[4], &sequence) != nil || json.Unmarshal(args[5], &payload) != nil || json.Unmarshal(args[6], &timeout) != nil {
		return nil, errors.New("argumentos de ws_send invalidos")
	}
	return protocol.WebSocketSend(protocol.WebSocketSendRequest{BaseURL: base, Path: path, DeviceID: device, Token: token, Sequence: sequence, Payload: payload, Timeout: time.Duration(timeout) * time.Second})
}

func activate(args []json.RawMessage) (any, error) {
	if len(args) != 6 {
		return nil, errors.New("activate requiere 6 argumentos")
	}
	var base, path, code, hardwareID, name string
	var timeout int
	if json.Unmarshal(args[0], &base) != nil || json.Unmarshal(args[1], &path) != nil || json.Unmarshal(args[2], &code) != nil || json.Unmarshal(args[3], &hardwareID) != nil || json.Unmarshal(args[4], &name) != nil || json.Unmarshal(args[5], &timeout) != nil {
		return nil, errors.New("argumentos de activate invalidos")
	}
	return protocol.Activate(base, path, code, hardwareID, name, time.Duration(timeout)*time.Second)
}

func weatherCurrent(args []json.RawMessage) (any, error) {
	if len(args) != 3 {
		return nil, errors.New("weather_current requiere latitud, longitud y timeout")
	}
	var latitude, longitude float64
	var timeout int
	if json.Unmarshal(args[0], &latitude) != nil || json.Unmarshal(args[1], &longitude) != nil || json.Unmarshal(args[2], &timeout) != nil {
		return nil, errors.New("argumentos de weather_current invalidos")
	}
	return protocol.WeatherCurrent(latitude, longitude, time.Duration(timeout)*time.Second)
}

func sign(args []json.RawMessage) (any, error) {
	if len(args) != 5 {
		return nil, errors.New("sign requiere token, timestamp, sequence, nonce, payload")
	}
	token, timestamp, sequence, nonce, payload, err := signatureArgs(args[:5])
	if err != nil {
		return nil, err
	}
	return protocol.Signature(token, timestamp, sequence, nonce, payload)
}
func verify(args []json.RawMessage) (any, error) {
	if len(args) != 7 {
		return nil, errors.New("verify requiere 7 argumentos")
	}
	token, timestamp, sequence, nonce, payload, err := signatureArgs(args[:5])
	if err != nil {
		return nil, err
	}
	var signature string
	var skew int
	if json.Unmarshal(args[5], &signature) != nil || json.Unmarshal(args[6], &skew) != nil || skew < 1 || skew > 3600 {
		return nil, errors.New("firma o ventana invalida")
	}
	if err := protocol.Verify(token, timestamp, sequence, nonce, payload, signature, time.Duration(skew)*time.Second, time.Now()); err != nil {
		return map[string]any{"valid": false, "error": err.Error()}, nil
	}
	return map[string]any{"valid": true}, nil
}

func verifyHash(args []json.RawMessage) (any, error) {
	if len(args) != 7 {
		return nil, errors.New("verify_hash requiere 7 argumentos")
	}
	key, timestamp, sequence, nonce, payload, err := signatureArgs(args[:5])
	if err != nil {
		return nil, err
	}
	var signature string
	var skew int
	if json.Unmarshal(args[5], &signature) != nil || json.Unmarshal(args[6], &skew) != nil || skew < 1 || skew > 3600 {
		return nil, errors.New("firma o ventana invalida")
	}
	if err := protocol.VerifyHash(key, timestamp, sequence, nonce, payload, signature, time.Duration(skew)*time.Second, time.Now()); err != nil {
		return map[string]any{"valid": false, "error": err.Error()}, nil
	}
	return map[string]any{"valid": true}, nil
}
func send(args []json.RawMessage) (any, error) {
	if len(args) != 7 {
		return nil, errors.New("send requiere 7 argumentos")
	}
	var base, path, device, token string
	var sequence int64
	var payload any
	var timeout int
	if json.Unmarshal(args[0], &base) != nil || json.Unmarshal(args[1], &path) != nil || json.Unmarshal(args[2], &device) != nil || json.Unmarshal(args[3], &token) != nil || json.Unmarshal(args[4], &sequence) != nil || json.Unmarshal(args[5], &payload) != nil || json.Unmarshal(args[6], &timeout) != nil {
		return nil, errors.New("argumentos de send invalidos")
	}
	return protocol.Send(protocol.SendRequest{BaseURL: base, Path: path, DeviceID: device, Token: token, Sequence: sequence, Payload: payload, Timeout: time.Duration(timeout) * time.Second})
}
func signatureArgs(args []json.RawMessage) (string, string, int64, string, any, error) {
	var token, timestamp, nonce string
	var sequence int64
	var payload any
	if json.Unmarshal(args[0], &token) != nil || json.Unmarshal(args[1], &timestamp) != nil || json.Unmarshal(args[2], &sequence) != nil || json.Unmarshal(args[3], &nonce) != nil || json.Unmarshal(args[4], &payload) != nil {
		return "", "", 0, "", nil, errors.New("argumentos de firma invalidos")
	}
	return token, timestamp, sequence, nonce, payload, nil
}
