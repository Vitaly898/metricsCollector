package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func Compute(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
