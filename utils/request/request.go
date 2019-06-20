package request

import (
	"crypto/hmac"
	"crypto/sha1"
	"errors"
	"fmt"
	"net/http"
)

// 检查 Header 必要参数是否存在
func CheckHeaderRequestEmpty(header http.Header) error {

	if header.Get("X-Coding-Event") == "" {
		return errors.New("request header event is empty")
	}

	if header.Get("X-Coding-Signature") == "" {
		return errors.New("request header signature is empty")
	}

	if header.Get("X-Coding-Delivery") == "" {
		return errors.New("request header delivery is empty")
	}

	return nil
}

// 检查签名
func CheckSign(body *[]byte, bodyMAC, key string) bool {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write(*body)
	expectedMAC := mac.Sum(nil)

	return hmac.Equal([]byte(bodyMAC), []byte(fmt.Sprintf("%x", expectedMAC)))
}
