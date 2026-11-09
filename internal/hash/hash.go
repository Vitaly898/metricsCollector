// Package hash вычисляет подписи HMAC-SHA256 для тел запросов и ответов.
package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// Compute возвращает HMAC-SHA256 данных data с ключом key в кодировке Base64.
func Compute(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
