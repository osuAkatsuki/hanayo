package oauthstate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	rawState, err := New("discord", 123, "secret", time.Unix(200, 0))
	if err != nil {
		t.Fatal(err)
	}

	parts := strings.Split(rawState, ".")
	if len(parts) != 3 || parts[0] != "v1" {
		t.Fatalf("invalid state format: %s", rawState)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}

	var state OAuthState
	if err := json.Unmarshal(payloadBytes, &state); err != nil {
		t.Fatal(err)
	}

	if state.Provider != "discord" || state.UserID != 123 || state.ExpiresAt != 200 || state.Nonce == "" {
		t.Fatalf("unexpected state payload: %+v", state)
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatal(err)
	}

	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		t.Fatal("state signature did not match")
	}
}
