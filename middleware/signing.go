package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
)

const HEADER_SIGNATURE = "X-Skill-Signature-256"

func NewSigning(signingKeys []string) (func(http.Handler) http.Handler, error) {
	if len(signingKeys) == 0 {
		return nil, fmt.Errorf("no signing keys provided")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			signature := request.Header.Get(HEADER_SIGNATURE)
			if signature == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			body, _ := io.ReadAll(request.Body)
			// Replace the body with a new reader after reading from the original
			request.Body = io.NopCloser(bytes.NewBuffer(body))

			if !signatureIsValid(signature, signingKeys, body) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, request)
		})
	}, nil
}

func signatureIsValid(signature string, signingKeys []string, body []byte) bool {
	for _, key := range signingKeys {
		hmac := hmac.New(sha256.New, []byte(key))

		// compute the HMAC
		hmac.Write(body)
		dataHmac := hmac.Sum(nil)

		hmacHex := hex.EncodeToString(dataHmac)

		if hmacHex == signature {
			return true
		}

	}
	return false
}
