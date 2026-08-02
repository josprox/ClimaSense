package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func telemetrySignature(token, timestamp string, sequence int64, nonce string, payload any) string {
	body, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	tokenHash := sha256.Sum256([]byte(token))
	payloadHash := sha256.Sum256(body)
	canonical := fmt.Sprintf("%s\n%d\n%s\n%x", timestamp, sequence, nonce, payloadHash)
	mac := hmac.New(sha256.New, []byte(fmt.Sprintf("%x", tokenHash)))
	_, _ = mac.Write([]byte(canonical))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

type response struct {
	Token          string `json:"token"`
	Type           string `json:"type"`
	Resource       string `json:"resource"`
	ActivationCode string `json:"activation_code"`
	Organization   struct {
		ID int64 `json:"id"`
	} `json:"organization"`
}

func request(client *http.Client, method, base, path string, payload any) (int, response, error) {
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return 0, response{}, err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, base+path, body)
	if err != nil {
		return 0, response{}, err
	}
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := client.Do(req)
	if err != nil {
		return 0, response{}, err
	}
	defer res.Body.Close()
	var decoded response
	_ = json.NewDecoder(res.Body).Decode(&decoded)
	return res.StatusCode, decoded, nil
}

func requireStatus(got int, allowed ...int) error {
	for _, status := range allowed {
		if got == status {
			return nil
		}
	}
	return fmt.Errorf("HTTP %d; se esperaba uno de %v", got, allowed)
}

func main() {
	base := os.Getenv("CLIMASENSE_TEST_BASE_URL")
	if base == "" {
		base = "http://127.0.0.1:18080"
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	client := &http.Client{Jar: jar, Timeout: 10 * time.Second}

	status, _, err := request(client, http.MethodPost, base, "/api/v1/setup/admin", map[string]any{
		"bootstrap_key": "test-only-bootstrap-key-with-at-least-32-characters",
		"email":         "live-admin@climasense.test",
		"password":      "LiveTest-Password-2026",
		"name":          "Live Test Admin",
	})
	if err != nil {
		panic(err)
	}
	if err = requireStatus(status, http.StatusCreated, http.StatusConflict); err != nil {
		panic(err)
	}

	status, _, err = request(client, http.MethodPost, base, "/api/v1/auth/login", map[string]any{
		"email": "live-admin@climasense.test", "password": "LiveTest-Password-2026",
	})
	if err != nil {
		panic(err)
	}
	if err = requireStatus(status, http.StatusOK); err != nil {
		panic(err)
	}

	status, token, err := request(client, http.MethodGet, base, "/api/v1/auth/ws-token", nil)
	if err != nil {
		panic(err)
	}
	if err = requireStatus(status, http.StatusOK); err != nil || token.Token == "" {
		panic("token WebSocket no disponible")
	}

	wsURL := strings.Replace(base, "http", "ws", 1) + "/ws/dashboard"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err = conn.WriteJSON(map[string]any{"type": "authenticate", "token": token.Token}); err != nil {
		panic(err)
	}
	var message response
	if err = conn.ReadJSON(&message); err != nil || message.Type != "ready" {
		panic("el panel no completo la autenticacion WebSocket")
	}

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	status, organization, err := request(client, http.MethodPost, base, "/api/v1/admin/organizations", map[string]any{
		"name":     "Live Test Organization",
		"slug":     "live-test-" + suffix,
		"email":    "live-client-" + suffix + "@climasense.test",
		"password": "LiveClient-Password-2026",
		"plan":     "starter",
	})
	if err != nil {
		panic(err)
	}
	if err = requireStatus(status, http.StatusCreated); err != nil {
		panic(err)
	}

	if err = conn.ReadJSON(&message); err != nil || message.Type != "refresh" || message.Resource != "organization" {
		panic("la mutacion HTTP no produjo la actualizacion WebSocket esperada")
	}

	status, activation, err := request(client, http.MethodPost, base, "/api/v1/admin/activation-codes", map[string]any{
		"organization_id": organization.Organization.ID, "label": "Concurrent activation test",
	})
	if err != nil {
		panic(err)
	}
	if err = requireStatus(status, http.StatusCreated); err != nil || activation.ActivationCode == "" {
		panic("no se pudo crear el codigo de activacion de prueba")
	}
	if err = conn.ReadJSON(&message); err != nil || message.Resource != "activation_code" {
		panic("la creacion del codigo no produjo evento en vivo")
	}

	results := make(chan int, 2)
	for index := 0; index < 2; index++ {
		go func(candidate int) {
			code, _, callErr := request(client, http.MethodPost, base, "/api/v1/edge/activate", map[string]any{
				"activation_code": activation.ActivationCode,
				"hardware_id":     fmt.Sprintf("live-hardware-%s-%d", suffix, candidate),
				"name":            "Concurrent activation test",
			})
			if callErr != nil {
				results <- 0
				return
			}
			results <- code
		}(index)
	}
	created := 0
	for index := 0; index < 2; index++ {
		if <-results == http.StatusCreated {
			created++
		}
	}
	if created != 1 {
		panic(fmt.Sprintf("el codigo de un solo uso produjo %d activaciones", created))
	}
	if err = conn.ReadJSON(&message); err != nil || message.Resource != "device_activation" {
		panic("la activacion no produjo evento en vivo")
	}

	measurementTime := time.Now().UTC().Truncate(time.Second)
	payload := map[string]any{
		"schema_version": 1,
		"measurements": []any{map[string]any{
			"schema_version":   1,
			"device_id":        "context-test-device",
			"sequence":         2,
			"measured_at":      measurementTime.Format(time.RFC3339),
			"temperature_c":    23.5,
			"pressure_pa":      101200,
			"pressure_hpa":     1012,
			"sensor_type":      "bmp280",
			"sensor_address":   "0x76",
			"sensor_status":    "ok",
			"firmware_version": "test",
			"runtime_version":  "test",
		}},
	}
	timestamp := measurementTime.Format(time.RFC3339)
	nonce := fmt.Sprintf("live-telemetry-%s", suffix)
	edgeURL := strings.Replace(base, "http", "ws", 1) + "/ws/edge"
	edgeConn, _, err := websocket.DefaultDialer.Dial(edgeURL, nil)
	if err != nil {
		panic(err)
	}
	defer edgeConn.Close()
	_ = edgeConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err = edgeConn.WriteJSON(map[string]any{
		"type":      "telemetry_batch",
		"device_id": "context-test-device",
		"timestamp": timestamp,
		"sequence":  2,
		"nonce":     nonce,
		"signature": telemetrySignature("context-test-device-token-2026", timestamp, 2, nonce, payload),
		"payload":   payload,
	}); err != nil {
		panic(err)
	}
	var ack map[string]any
	if err = edgeConn.ReadJSON(&ack); err != nil || ack["ok"] != true || ack["last_seen_at"] == nil {
		panic(fmt.Sprintf("la telemetria WebSocket no actualizo last_seen_at: %v", ack))
	}
	if err = conn.ReadJSON(&message); err != nil || message.Type != "refresh" || message.Resource != "telemetry" {
		panic("la telemetria no produjo actualizacion inmediata del panel")
	}
	fmt.Println("server-dashboard-live-ok")
}
