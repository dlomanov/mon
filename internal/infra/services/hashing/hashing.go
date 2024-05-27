package hashing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

const (
	HeaderHash = "HashSHA256"
)

// ComputeBase64URLHash computes a Base64 URL-encoded hash of the given value using HMAC-SHA256
// with the provided secret key. The resulting hash is suitable for use in URLs and other contexts
// where a URL-safe string is required.
func ComputeBase64URLHash(key string, value []byte) string {
	hf := hmac.New(sha256.New, []byte(key))
	hf.Write(value)
	return base64.URLEncoding.EncodeToString(hf.Sum(nil))
}
