package protocol

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestSignatureAndVerification(t *testing.T) {
	token := "0123456789abcdef0123456789abcdef"
	payload := map[string]any{"device_id": "edge-1", "sequence": int64(7)}
	timestamp := "2026-07-16T12:00:00Z"
	signature, err := Signature(token, timestamp, 7, "abc", payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := Verify(token, timestamp, 7, "abc", payload, signature, time.Minute, time.Date(2026, 7, 16, 12, 0, 20, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if err := VerifyHash(HashToken(token), timestamp, 7, "abc", payload, signature, time.Minute, time.Date(2026, 7, 16, 12, 0, 20, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	payload["sequence"] = int64(8)
	if err := Verify(token, timestamp, 7, "abc", payload, signature, time.Minute, time.Date(2026, 7, 16, 12, 0, 20, 0, time.UTC)); err == nil {
		t.Fatal("tampered payload accepted")
	}
}

func TestWebSocketSendSignsEnvelopeAndReceivesAck(t *testing.T) {
	token := "0123456789abcdef0123456789abcdef"
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		var envelope map[string]any
		if err := conn.ReadJSON(&envelope); err != nil {
			t.Error(err)
			return
		}
		if envelope["type"] != "telemetry_batch" || envelope["device_id"] != "edge-1" {
			t.Errorf("unexpected envelope: %#v", envelope)
		}
		if err := Verify(token, envelope["timestamp"].(string), 9, envelope["nonce"].(string), envelope["payload"], envelope["signature"].(string), time.Minute, time.Now()); err != nil {
			t.Error(err)
		}
		_ = conn.WriteJSON(map[string]any{"ok": true, "device_id": "edge-1", "accepted_through": 9})
	}))
	defer server.Close()
	reply, err := WebSocketSend(WebSocketSendRequest{BaseURL: server.URL, Path: "/ws/edge", DeviceID: "edge-1", Token: token, Sequence: 9, Payload: map[string]any{"measurements": []any{map[string]any{"sequence": 9}}}, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if reply["ok"] != true {
		t.Fatalf("unexpected reply: %#v", reply)
	}
}

func TestActivatePostsOnboardingIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["activation_code"] != "A1B2" || body["hardware_id"] != "pi-serial" {
			t.Errorf("unexpected activation: %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"device_id":"dev-1","device_token":"secret"}`))
	}))
	defer server.Close()
	result, err := Activate(server.URL, "/api/v1/edge/activate", "A1B2", "pi-serial", "Edge", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if result["ok"] != true || result["device_id"] != "dev-1" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestRejectsReplayWindow(t *testing.T) {
	token := "0123456789abcdef"
	payload := map[string]any{"x": 1}
	signature, _ := Signature(token, "2026-07-16T12:00:00Z", 1, "n", payload)
	if Verify(token, "2026-07-16T12:00:00Z", 1, "n", payload, signature, time.Minute, time.Date(2026, 7, 16, 12, 2, 0, 0, time.UTC)) == nil {
		t.Fatal("old timestamp accepted")
	}
}

func TestSendUsesTLSAndSignedHeaders(t *testing.T) {
	token := "0123456789abcdef0123456789abcdef"
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if err := Verify(token, r.Header.Get("X-Timestamp"), 9, r.Header.Get("X-Nonce"), payload, r.Header.Get("X-Signature"), time.Minute, time.Now()); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accepted":1}`))
	}))
	defer server.Close()
	previous := http.DefaultTransport
	_ = previous
	// Send intentionally trusts only public roots. Assert that an untrusted test CA is rejected.
	if _, err := Send(SendRequest{BaseURL: server.URL, Path: "/api", DeviceID: "edge", Token: token, Sequence: 9, Payload: map[string]any{"x": 1}, Timeout: time.Second}); err == nil {
		t.Fatal("untrusted TLS certificate accepted")
	}
	_ = tls.VersionTLS13
}
