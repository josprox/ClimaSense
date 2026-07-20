package protocol

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func CanonicalPayload(payload any) ([]byte, error) { return json.Marshal(payload) }

func Signature(token, timestamp string, sequence int64, nonce string, payload any) (string, error) {
	if len(token) < 16 {
		return "", errors.New("token demasiado corto")
	}
	return SignatureWithKey(HashToken(token), timestamp, sequence, nonce, payload)
}

func SignatureWithKey(key, timestamp string, sequence int64, nonce string, payload any) (string, error) {
	if len(key) != 64 {
		return "", errors.New("clave HMAC derivada invalida")
	}
	body, err := CanonicalPayload(payload)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(body)
	canonical := timestamp + "\n" + strconv.FormatInt(sequence, 10) + "\n" + nonce + "\n" + hex.EncodeToString(digest[:])
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(canonical))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func VerifyHash(tokenHash, timestamp string, sequence int64, nonce string, payload any, signature string, maxSkew time.Duration, now time.Time) error {
	parsed, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return errors.New("timestamp debe usar RFC3339")
	}
	delta := now.UTC().Sub(parsed.UTC())
	if delta < 0 {
		delta = -delta
	}
	if delta > maxSkew {
		return fmt.Errorf("timestamp fuera de ventana: %s", delta.Round(time.Second))
	}
	expected, err := SignatureWithKey(tokenHash, timestamp, sequence, nonce, payload)
	if err != nil {
		return err
	}
	expectedBytes, err1 := hex.DecodeString(expected)
	providedBytes, err2 := hex.DecodeString(signature)
	if err1 != nil || err2 != nil || !hmac.Equal(expectedBytes, providedBytes) {
		return errors.New("firma HMAC invalida")
	}
	return nil
}

func Verify(token, timestamp string, sequence int64, nonce string, payload any, signature string, maxSkew time.Duration, now time.Time) error {
	parsed, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return errors.New("timestamp debe usar RFC3339")
	}
	delta := now.UTC().Sub(parsed.UTC())
	if delta < 0 {
		delta = -delta
	}
	if delta > maxSkew {
		return fmt.Errorf("timestamp fuera de ventana: %s", delta.Round(time.Second))
	}
	expected, err := Signature(token, timestamp, sequence, nonce, payload)
	if err != nil {
		return err
	}
	expectedBytes, err1 := hex.DecodeString(expected)
	providedBytes, err2 := hex.DecodeString(signature)
	if err1 != nil || err2 != nil || !hmac.Equal(expectedBytes, providedBytes) {
		return errors.New("firma HMAC invalida")
	}
	return nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func RandomToken(size int) (string, error) {
	if size < 16 || size > 128 {
		return "", errors.New("tamano de token fuera de rango")
	}
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

type SendRequest struct {
	BaseURL, Path, DeviceID, Token string
	Sequence                       int64
	Payload                        any
	Timeout                        time.Duration
}

func Send(input SendRequest) (map[string]any, error) {
	base, err := url.Parse(input.BaseURL)
	if err != nil || (base.Scheme != "https" && base.Scheme != "http") || base.Host == "" {
		return nil, errors.New("base_url debe ser una URL HTTP(S) absoluta")
	}
	relative, err := url.Parse(input.Path)
	if err != nil || relative.IsAbs() || !strings.HasPrefix(relative.Path, "/") {
		return nil, errors.New("path HTTPS invalido")
	}
	target := base.ResolveReference(relative)
	body, err := CanonicalPayload(input.Payload)
	if err != nil {
		return nil, err
	}
	timestamp := time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
	nonce, err := RandomToken(16)
	if err != nil {
		return nil, err
	}
	signature, err := Signature(input.Token, timestamp, input.Sequence, nonce, input.Payload)
	if err != nil {
		return nil, err
	}
	if input.Timeout <= 0 || input.Timeout > 60*time.Second {
		input.Timeout = 15 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), input.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ClimaSense-Edge/0.1")
	req.Header.Set("X-Device-ID", input.DeviceID)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Sequence", strconv.FormatInt(input.Sequence, 10))
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-Signature", signature)
	client := &http.Client{Timeout: input.Timeout, Transport: &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}, TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: input.Timeout}}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTPS: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	result := map[string]any{"status": resp.StatusCode, "ok": resp.StatusCode >= 200 && resp.StatusCode < 300}
	if len(responseBody) > 0 {
		var decoded any
		if json.Unmarshal(responseBody, &decoded) == nil {
			result["body"] = decoded
		} else {
			result["body_text"] = string(responseBody)
		}
	}
	return result, nil
}

type WebSocketSendRequest struct {
	BaseURL, Path, DeviceID, Token string
	Sequence                       int64
	Payload                        any
	Timeout                        time.Duration
}

func WebSocketSend(input WebSocketSendRequest) (map[string]any, error) {
	base, err := url.Parse(input.BaseURL)
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https" && base.Scheme != "ws" && base.Scheme != "wss") {
		return nil, errors.New("base_url WebSocket invalida")
	}
	if base.Scheme == "http" {
		base.Scheme = "ws"
	} else if base.Scheme == "https" {
		base.Scheme = "wss"
	}
	relative, err := url.Parse(input.Path)
	if err != nil || relative.IsAbs() || !strings.HasPrefix(relative.Path, "/") {
		return nil, errors.New("path WebSocket invalido")
	}
	target := base.ResolveReference(relative)
	if input.Timeout <= 0 || input.Timeout > 60*time.Second {
		input.Timeout = 15 * time.Second
	}
	timestamp := time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
	nonce, err := RandomToken(16)
	if err != nil {
		return nil, err
	}
	signature, err := Signature(input.Token, timestamp, input.Sequence, nonce, input.Payload)
	if err != nil {
		return nil, err
	}
	dialer := websocket.Dialer{HandshakeTimeout: input.Timeout, TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
	conn, _, err := dialer.Dial(target.String(), http.Header{"User-Agent": []string{"ClimaSense-Edge/0.2"}})
	if err != nil {
		return nil, fmt.Errorf("WebSocket: %w", err)
	}
	defer conn.Close()
	deadline := time.Now().Add(input.Timeout)
	_ = conn.SetWriteDeadline(deadline)
	envelope := map[string]any{
		"type":      "telemetry_batch",
		"device_id": input.DeviceID,
		"timestamp": timestamp,
		"sequence":  input.Sequence,
		"nonce":     nonce,
		"signature": signature,
		"payload":   input.Payload,
	}
	if err := conn.WriteJSON(envelope); err != nil {
		return nil, fmt.Errorf("WebSocket write: %w", err)
	}
	_ = conn.SetReadDeadline(deadline)
	for {
		var reply map[string]any
		if err := conn.ReadJSON(&reply); err != nil {
			return nil, fmt.Errorf("WebSocket read: %w", err)
		}
		if reply["device_id"] == input.DeviceID {
			return reply, nil
		}
	}
}

func Activate(baseURL, path, activationCode, hardwareID, name string, timeout time.Duration) (map[string]any, error) {
	base, err := url.Parse(baseURL)
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") {
		return nil, errors.New("base_url de activacion invalida")
	}
	relative, err := url.Parse(path)
	if err != nil || relative.IsAbs() || !strings.HasPrefix(relative.Path, "/") {
		return nil, errors.New("path de activacion invalido")
	}
	if timeout <= 0 || timeout > 60*time.Second {
		timeout = 15 * time.Second
	}
	body, err := json.Marshal(map[string]any{"activation_code": activationCode, "hardware_id": hardwareID, "name": name})
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base.ResolveReference(relative).String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ClimaSense-Onboarding/0.2")
	resp, err := (&http.Client{Timeout: timeout, Transport: &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}}}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("activacion: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	result := map[string]any{"ok": resp.StatusCode >= 200 && resp.StatusCode < 300, "status": resp.StatusCode}
	var decoded map[string]any
	if json.Unmarshal(data, &decoded) == nil {
		for key, value := range decoded {
			result[key] = value
		}
	}
	return result, nil
}

func WeatherCurrent(latitude, longitude float64, timeout time.Duration) (map[string]any, error) {
	if latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		return nil, errors.New("coordenadas fuera de rango")
	}
	if timeout <= 0 || timeout > 30*time.Second {
		timeout = 10 * time.Second
	}
	endpoint, _ := url.Parse("https://api.open-meteo.com/v1/forecast")
	query := endpoint.Query()
	query.Set("latitude", strconv.FormatFloat(latitude, 'f', 7, 64))
	query.Set("longitude", strconv.FormatFloat(longitude, 'f', 7, 64))
	query.Set("current", "temperature_2m,relative_humidity_2m,apparent_temperature,surface_pressure")
	query.Set("timezone", "auto")
	endpoint.RawQuery = query.Encode()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	req.Header.Set("User-Agent", "ClimaSense-Server/0.2")
	resp, err := (&http.Client{Timeout: timeout, Transport: &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}}}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("Open-Meteo: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Open-Meteo HTTP %d", resp.StatusCode)
	}
	var decoded struct {
		Current map[string]any `json:"current"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&decoded); err != nil {
		return nil, err
	}
	return map[string]any{
		"temperature_c":          decoded.Current["temperature_2m"],
		"apparent_temperature_c": decoded.Current["apparent_temperature"],
		"relative_humidity":      decoded.Current["relative_humidity_2m"],
		"surface_pressure_hpa":   decoded.Current["surface_pressure"],
		"observed_at":            decoded.Current["time"],
		"source":                 "open-meteo",
	}, nil
}
