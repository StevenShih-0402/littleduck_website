// backend/internal/service/payment/checkmac_test.go
// CheckMacValue 單元測試：以綠界官方文件範例驗證，並涵蓋特殊字元、排序與驗證失敗

package payment

import (
	"strings"
	"testing"
)

// 綠界官方文件「檢查碼機制」的 AIO 範例（SHA256）
const (
	docHashKey = "pwFHCqoQZGmho4w6"
	docHashIV  = "EkRm7iFT261dpevs"
	docWantMac = "6C51C9E6888DE861FD62FB1DD17029FC742634498FD813DC43D4243B5685B840"
)

func docParams() map[string]string {
	return map[string]string{
		"MerchantID":        "3002607",
		"MerchantTradeNo":   "ecpay20230312153023",
		"MerchantTradeDate": "2023/03/12 15:30:23",
		"PaymentType":       "aio",
		"TotalAmount":       "30000",
		"TradeDesc":         "促銷方案",
		"ItemName":          "Apple iPhone 15",
		"ReturnURL":         "https://www.ecpay.com.tw/receive.php",
		"ChoosePayment":     "ALL",
		"EncryptType":       "1",
	}
}

func TestCheckMacValue_MatchesOfficialDocExample(t *testing.T) {
	// Arrange
	params := docParams()

	// Act
	got := CheckMacValue(params, docHashKey, docHashIV)

	// Assert
	if got != docWantMac {
		t.Errorf("CheckMacValue = %s, want %s", got, docWantMac)
	}
}

func TestCheckMacValue_IgnoresExistingCheckMacField(t *testing.T) {
	// Arrange
	params := docParams()
	params["CheckMacValue"] = "STALE"

	// Act
	got := CheckMacValue(params, docHashKey, docHashIV)

	// Assert
	if got != docWantMac {
		t.Errorf("既有 CheckMacValue 不應參與計算，得到 %s", got)
	}
}

func TestCheckMacValue_SortsKeysCaseInsensitively(t *testing.T) {
	// Arrange：不分大小寫排序時 a 在 B 之前；byte 排序則 B 在 a 之前
	params := map[string]string{"B": "2", "a": "1"}

	// Act
	got := CheckMacValue(params, "k", "i")

	// Assert
	want := sha256Upper("hashkey%3dk%26a%3d1%26b%3d2%26hashiv%3di")
	if got != want {
		t.Errorf("應依 a、B 順序串接，CheckMacValue = %s, want %s", got, want)
	}
	if got == sha256Upper("hashkey%3dk%26b%3d2%26a%3d1%26hashiv%3di") {
		t.Error("不應使用 byte 排序")
	}
}

func TestCheckMacValue_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string // 期望進入雜湊的 URL encode 結果中該值的片段
	}{
		{"空白轉加號", "a b", "a+b"},
		{"連字號保留", "a-b", "a-b"},
		{"底線保留", "a_b", "a_b"},
		{"句點保留", "a.b", "a.b"},
		{"驚嘆號保留", "a!b", "a!b"},
		{"星號保留", "a*b", "a*b"},
		{"左右括號保留", "(a)", "(a)"},
		{"斜線編碼並轉小寫", "a/b", "a%2fb"},
		{"井字號編碼", "a#b", "a%23b"},
		{"中文以小寫百分比編碼", "鴨", "%e9%b4%a8"},
		{"波浪號編碼", "a~b", "a%7eb"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			params := map[string]string{"X": tc.value}
			wantRaw := "hashkey%3dk%26x%3d" + tc.want + "%26hashiv%3di"

			// Act
			got := CheckMacValue(params, "k", "i")

			// Assert
			if want := sha256Upper(wantRaw); got != want {
				t.Errorf("value %q：CheckMacValue 與預期 encode %q 的雜湊不符", tc.value, wantRaw)
			}
		})
	}
}

func TestVerifyCheckMac(t *testing.T) {
	signed := docParams()
	signed["CheckMacValue"] = docWantMac

	lower := docParams()
	lower["CheckMacValue"] = strings.ToLower(docWantMac)

	tampered := docParams()
	tampered["CheckMacValue"] = docWantMac
	tampered["TotalAmount"] = "1"

	wrongMac := docParams()
	wrongMac["CheckMacValue"] = strings.Repeat("A", 64)

	missing := docParams()

	empty := docParams()
	empty["CheckMacValue"] = ""

	tests := []struct {
		name   string
		params map[string]string
		key    string
		want   bool
	}{
		{"正確", signed, docHashKey, true},
		{"小寫檢查碼也接受", lower, docHashKey, true},
		{"參數被竄改", tampered, docHashKey, false},
		{"檢查碼錯誤", wrongMac, docHashKey, false},
		{"HashKey 不同", signed, "other", false},
		{"缺少檢查碼", missing, docHashKey, false},
		{"檢查碼為空", empty, docHashKey, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := VerifyCheckMac(tc.params, tc.key, docHashIV); got != tc.want {
				t.Errorf("VerifyCheckMac = %v, want %v", got, tc.want)
			}
		})
	}
}
