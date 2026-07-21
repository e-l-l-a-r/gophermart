package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func GetAuthKey(login string, password string) string {
	h := hmac.New(sha256.New, []byte(password))
	if _, err := h.Write([]byte(login)); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}
