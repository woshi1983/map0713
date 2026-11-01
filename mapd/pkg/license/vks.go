// Package license verifies offline VKS Supervisor licenses.
package license

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const MaxTokenBytes = 8192

type Payload struct {
	Version       int      `json:"version"`
	SupervisorUID string   `json:"supervisorUID"`
	ExpiresAt     string   `json:"expiresAt"`
	Features      []string `json:"features"`
}

// PublicKey is configured on mapd. The matching private key stays with the issuer.
func PublicKey() (ed25519.PublicKey, error) {
	encoded := os.Getenv("MAP_VKS_LICENSE_PUBLIC_KEY")
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return nil, errors.New("MAP_VKS_LICENSE_PUBLIC_KEY must be a base64 Ed25519 public key")
	}
	return ed25519.PublicKey(key), nil
}

// VerifyVKS verifies the exact signed bytes before decoding the payload.
func VerifyVKS(token string, key ed25519.PublicKey, supervisorUID string, now time.Time) (*Payload, error) {
	if len(token) == 0 || len(token) > MaxTokenBytes || len(key) != ed25519.PublicKeySize {
		return nil, errors.New("invalid VKS license")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid VKS license format")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !ed25519.Verify(key, []byte(parts[0]), sig) {
		return nil, errors.New("invalid VKS license signature")
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("invalid VKS license payload")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var p Payload
	if err := decoder.Decode(&p); err != nil {
		return nil, errors.New("invalid VKS license payload")
	}
	if err := decoder.Decode(new(interface{})); err != io.EOF {
		return nil, errors.New("invalid VKS license payload")
	}
	if p.Version != 1 || p.SupervisorUID == "" || supervisorUID == "" || p.SupervisorUID != supervisorUID {
		return nil, errors.New("VKS license does not match this Supervisor")
	}
	allowed := false
	for _, feature := range p.Features {
		if feature == "vks" {
			allowed = true
		}
	}
	if !allowed {
		return nil, errors.New("VKS feature is not licensed")
	}
	expiry, err := time.Parse(time.RFC3339, p.ExpiresAt)
	if err != nil || !now.Before(expiry) {
		return nil, fmt.Errorf("VKS license is expired or has invalid expiry")
	}
	return &p, nil
}
