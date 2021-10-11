package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"os"
)

func HashPassword(password string, secret string) string {
	secret = os.Getenv("SC_KEY") + secret
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(password))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
