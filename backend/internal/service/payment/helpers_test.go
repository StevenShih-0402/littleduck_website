// backend/internal/service/payment/helpers_test.go
// 測試共用的小工具

package payment

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func sha256Upper(s string) string {
	sum := sha256.Sum256([]byte(s))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}
