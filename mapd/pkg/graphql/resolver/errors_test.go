package resolver

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

func TestIsObjectNotFound(t *testing.T) {
	if !isObjectNotFound(&gqlerror.Error{Extensions: map[string]interface{}{"code": fmt.Sprintf("0x%08X", CodeObjectNotFound)}}) {
		t.Fatal("object not found error was not recognized")
	}
	if isObjectNotFound(&gqlerror.Error{Extensions: map[string]interface{}{"code": fmt.Sprintf("0x%08X", CodeInvalidRequestError)}}) {
		t.Fatal("invalid request was treated as object not found")
	}
	if isObjectNotFound(errors.New("object not found")) {
		t.Fatal("plain error was treated as object not found")
	}
}

func TestGatewayErrorMessage(t *testing.T) {
	tests := []struct {
		name    string
		reason  string
		code    ErrCode
		message string
	}{
		{"kpack unavailable", "the server could not find the requested resource (get builds.kpack.io)", CodeInvalidRequestError, "source builds are not enabled on this cluster"},
		{"velero unavailable", "the server could not find the requested resource (get backups.velero.io)", CodeInvalidRequestError, "backup and restore are not enabled on this cluster"},
		{"other gateway failure", "dial tcp 172.16.84.10:6443: connect: connection refused", CodeGatewayError, "gateway error: upstream connection refused"},
		{"redact unknown failure", "https://user:password@supervisor.example/api?token=secret failed", CodeGatewayError, "gateway error: upstream request failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, message := gatewayErrorMessage(errors.New(tt.reason))
			if code != tt.code || message != tt.message {
				t.Fatalf("got code=%#x message=%q, want code=%#x message=%q", code, message, tt.code, tt.message)
			}
		})
	}
}

func TestGatewayErrorMessageNormalizesLongReason(t *testing.T) {
	_, message := gatewayErrorMessage(errors.New(strings.Repeat("x", 500) + "\nsecret"))
	if strings.Contains(message, "secret") || strings.Contains(message, "\n") {
		t.Fatalf("message exposes untrusted error text: %q", message)
	}
}
