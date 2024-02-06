package hashing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

const (
	HeaderHash = "HashSHA256"
)

func ComputeBase64URLHash(key string, value []byte) string {
	hf := hmac.New(sha256.New, []byte(key))
	hf.Write(value)
	return base64.URLEncoding.EncodeToString(hf.Sum(nil))
}
