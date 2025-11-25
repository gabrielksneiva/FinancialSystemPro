package metrics

import "testing"

func TestRecordTransactionMetrics(t *testing.T) {
	RecordDeposit(10, true)
	RecordWithdraw(5, false)
	RecordTransfer(3, true)
	RecordUserCreated()
	RecordAuthentication(true)
	RecordWalletCreated("tron")
	RecordBlockchainTransaction("tron", false)
}
