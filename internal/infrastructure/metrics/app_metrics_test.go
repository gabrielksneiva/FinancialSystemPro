package metrics

import (
	"testing"
	"time"
)

func TestMetricsRecordingAndSnapshot(t *testing.T) {
	start := time.Now().Add(-2 * time.Second)
	RecordDeposit()
	RecordWithdraw()
	RecordTransfer()
	RecordFailure()
	RecordRequestTime(15 * time.Millisecond)
	snap := Snapshot(start)
	tx := snap["transactions"].(map[string]interface{})
	api := snap["api"].(map[string]interface{})
	if tx["deposits"].(int64) < 1 || tx["withdraws"].(int64) < 1 || tx["transfers"].(int64) < 1 {
		t.Fatalf("contadores de transações não atualizados: %+v", tx)
	}
	if api["total_requests"].(int64) < 3 {
		t.Fatalf("total_requests inconsistente: %+v", api)
	}
}
