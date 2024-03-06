package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSignatureValidation(t *testing.T) {
	var (
		signingKey = "It's a Secret to Everybody"
		payload    = "Hello, World!"
	)

	if !signatureIsValid("757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17", []string{signingKey}, []byte(payload)) {
		t.Error("signature should be valid")
	}
}

func TestSignatureHandler(t *testing.T) {
	var (
		signingKey = "It's a Secret to Everybody"
		payload    = "Hello, World!"
		signature  = "757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17"
		rr         = httptest.NewRecorder()
	)

	// Create a signed request
	req := httptest.NewRequest("POST", "/", strings.NewReader(payload))
	req.Header.Set(HEADER_SIGNATURE, signature)

	// Create a handler that will validate the signature
	signingMiddleware, err := NewSigning([]string{signingKey})
	if err != nil {
		t.Error("error creating signing middleware")
	}

	// Create and call the handler
	handler := signingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	handler.ServeHTTP(rr, req)

	// Check the status code
	if rr.Code != 200 {
		t.Errorf("status code should be 200, got: %d", rr.Code)
	}
}
