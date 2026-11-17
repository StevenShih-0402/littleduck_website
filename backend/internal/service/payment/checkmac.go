// backend/internal/service/payment/checkmac.go
// 綠界 CheckMacValue（SHA256）計算與驗證：純函式，不碰設定與資料庫

package payment

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

// checkMacField 參與簽章時要排除的欄位名稱（CheckMacValue 本身）
const checkMacField = "CheckMacValue"

// dotNetUnescape 綠界規定 URL encode 後（轉小寫）要還原的字元，對應 .NET 不編碼的集合
var dotNetUnescape = strings.NewReplacer(
	"%2d", "-",
	"%5f", "_",
	"%2e", ".",
	"%21", "!",
	"%2a", "*",
	"%28", "(",
	"%29", ")",
)

// CheckMacValue 依綠界規則計算檢查碼：
// 參數（排除 CheckMacValue）依 key 不分大小寫排序 → 組成 k=v&… → 前後加 HashKey／HashIV
// → URL encode 並轉小寫 → 還原 .NET 不編碼字元 → SHA256 → 轉大寫。
func CheckMacValue(params map[string]string, hashKey, hashIV string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if strings.EqualFold(k, checkMacField) {
			continue
		}
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := strings.ToLower(keys[i]), strings.ToLower(keys[j])
		if a != b {
			return a < b
		}
		return keys[i] < keys[j]
	})

	var sb strings.Builder
	sb.WriteString("HashKey=")
	sb.WriteString(hashKey)
	for _, k := range keys {
		sb.WriteString("&")
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
	}
	sb.WriteString("&HashIV=")
	sb.WriteString(hashIV)

	// Go 的 QueryEscape 不編碼 ~，但綠界（.NET／PHP urlencode）會編成 %7e
	escaped := strings.ReplaceAll(url.QueryEscape(sb.String()), "~", "%7E")
	encoded := dotNetUnescape.Replace(strings.ToLower(escaped))
	sum := sha256.Sum256([]byte(encoded))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

// VerifyCheckMac 驗證 params 內的 CheckMacValue 是否與重新計算的值相符
// （不分大小寫、常數時間比較）；缺少 CheckMacValue 視為驗證失敗。
func VerifyCheckMac(params map[string]string, hashKey, hashIV string) bool {
	got := ""
	for k, v := range params {
		if strings.EqualFold(k, checkMacField) {
			got = v
			break
		}
	}
	if got == "" {
		return false
	}
	want := CheckMacValue(params, hashKey, hashIV)
	return subtle.ConstantTimeCompare([]byte(strings.ToUpper(got)), []byte(want)) == 1
}
