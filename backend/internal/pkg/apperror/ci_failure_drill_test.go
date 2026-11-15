// backend/internal/pkg/apperror/ci_failure_drill_test.go
// CI 失敗演練（SDP-117 5.4）：只在 Drone（DRONE=true）失敗，本機略過；不可合併

package apperror

import (
	"os"
	"testing"
)

func TestCIFailureDrill(t *testing.T) {
	if os.Getenv("DRONE") != "true" {
		t.Skip("非 Drone 環境，略過 CI 失敗演練")
	}
	t.Fatal("CI 失敗演練（SDP-117 5.4），不可合併")
}
