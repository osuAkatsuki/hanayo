package oauthstate

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"
)

type OAuthState struct {
	Provider  string `json:"provider"`
	UserID    int    `json:"user_id"`
	ExpiresAt int64  `json:"expires_at"`
	Nonce     string `json:"nonce"`
}

func New(provider string, userID int, secret string, expiresAt time.Time) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	payload, err := json.Marshal(OAuthState{
		Provider:  provider,
		UserID:    userID,
		ExpiresAt: expiresAt.Unix(),
		Nonce:     base64.RawURLEncoding.EncodeToString(nonce),
	})
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signedValue := "v1." + encodedPayload
	return signedValue + "." + base64.RawURLEncoding.EncodeToString(sign(signedValue, secret)), nil
}

func sign(value string, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(value))
	return mac.Sum(nil)
}
