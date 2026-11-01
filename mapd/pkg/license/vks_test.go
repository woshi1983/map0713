package license

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestVerifyVKS(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	uid := "supervisor-uid"
	payload := Payload{Version: 1, SupervisorUID: uid, ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339), Features: []string{"vks"}}
	body, _ := json.Marshal(payload)
	encoded := base64.RawURLEncoding.EncodeToString(body)
	token := encoded + "." + base64.RawURLEncoding.EncodeToString(ed25519.Sign(priv, []byte(encoded)))
	if _, err := VerifyVKS(token, pub, uid, time.Now()); err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}
	if _, err := VerifyVKS(token, pub, "another-supervisor", time.Now()); err == nil {
		t.Fatal("wrong supervisor accepted")
	}
	if _, err := VerifyVKS(token+"x", pub, uid, time.Now()); err == nil {
		t.Fatal("tampered signature accepted")
	}
	if _, err := VerifyVKS(token, pub, uid, time.Now().Add(2*time.Hour)); err == nil {
		t.Fatal("expired token accepted")
	}
	if _, err := VerifyVKS(strings.Repeat("x", MaxTokenBytes+1), pub, uid, time.Now()); err == nil {
		t.Fatal("oversized token accepted")
	}
}
