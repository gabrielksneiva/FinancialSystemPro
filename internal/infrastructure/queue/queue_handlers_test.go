package workers

import (
	"context"
	"encoding/json"
	repositories "financial-system-pro/internal/infrastructure/database"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupLiteDB cria sqlite simplificado para testar handlers.
func setupLiteDB(t *testing.T) *repositories.NewDatabase {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite open: %v", err)
	}
	// Cria tabelas mínimas.
	exec := db.Exec("CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT, password TEXT);")
	if exec.Error != nil {
		t.Fatalf("create users: %v", exec.Error)
	}
	if err := db.Exec("CREATE TABLE balances (id TEXT, user_id TEXT, amount NUMERIC, created_at DATETIME);").Error; err != nil {
		t.Fatalf("create balances: %v", err)
	}
	if err := db.Exec("CREATE TABLE transactions (id TEXT, account_id TEXT, amount NUMERIC, type TEXT, category TEXT, description TEXT, tron_tx_hash TEXT, tron_tx_status TEXT, onchain_tx_hash TEXT, onchain_tx_status TEXT, onchain_chain TEXT, created_at DATETIME);").Error; err != nil {
		t.Fatalf("create transactions: %v", err)
	}
	return &repositories.NewDatabase{DB: db, Logger: zap.NewNop()}
}

// helper para criar QM sem Redis real.
func newTestQM(db *repositories.NewDatabase) *QueueManager {
	return &QueueManager{logger: zap.NewNop(), database: db}
}

func TestHandleDeposit_Success(t *testing.T) {
	db := setupLiteDB(t)
	qm := newTestQM(db)
	uid := uuid.New()
	seedBalance(t, db, uid, decimal.NewFromFloat(0))
	payload := TransactionPayload{UserID: uid.String(), Amount: "10.50"}
	data, _ := json.Marshal(payload)
	task := asynq.NewTask(TypeDeposit, data)
	if err := qm.handleDeposit(context.Background(), task); err != nil {
		t.Fatalf("deposit erro: %v", err)
	}
}

func TestHandleDeposit_InvalidAmount(t *testing.T) {
	db := setupLiteDB(t)
	qm := newTestQM(db)
	uid := uuid.New().String()
	payload := TransactionPayload{UserID: uid, Amount: "0"}
	data, _ := json.Marshal(payload)
	task := asynq.NewTask(TypeDeposit, data)
	if err := qm.handleDeposit(context.Background(), task); err == nil {
		t.Fatalf("esperava erro amount zero")
	}
}

func seedBalance(t *testing.T, db *repositories.NewDatabase, userID uuid.UUID, amount decimal.Decimal) {
	if err := db.DB.Exec("INSERT INTO balances (id,user_id,amount,created_at) VALUES (?,?,?,CURRENT_TIMESTAMP)", uuid.New().String(), userID.String(), amount.String()).Error; err != nil {
		t.Fatalf("seed balance: %v", err)
	}
}

func seedUser(t *testing.T, db *repositories.NewDatabase, userID uuid.UUID, email string) {
	if err := db.DB.Exec("INSERT INTO users (id,email,password) VALUES (?,?,?)", userID.String(), email, "pwd").Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func TestHandleWithdraw_Insufficient(t *testing.T) {
	db := setupLiteDB(t)
	qm := newTestQM(db)
	uid := uuid.New()
	seedBalance(t, db, uid, decimal.NewFromFloat(5))
	payload := TransactionPayload{UserID: uid.String(), Amount: "10"}
	data, _ := json.Marshal(payload)
	task := asynq.NewTask(TypeWithdraw, data)
	if err := qm.handleWithdraw(context.Background(), task); err == nil {
		t.Fatalf("esperava erro insuficiente")
	}
}

func TestHandleWithdraw_Success(t *testing.T) {
	db := setupLiteDB(t)
	qm := newTestQM(db)
	uid := uuid.New()
	seedBalance(t, db, uid, decimal.NewFromFloat(15))
	payload := TransactionPayload{UserID: uid.String(), Amount: "10"}
	data, _ := json.Marshal(payload)
	task := asynq.NewTask(TypeWithdraw, data)
	if err := qm.handleWithdraw(context.Background(), task); err != nil {
		t.Fatalf("withdraw erro: %v", err)
	}
}

func TestHandleTransfer_RecipientNotFound(t *testing.T) {
	db := setupLiteDB(t)
	qm := newTestQM(db)
	sender := uuid.New()
	seedUser(t, db, sender, "sender@x.com")
	seedBalance(t, db, sender, decimal.NewFromFloat(20))
	payload := TransactionPayload{UserID: sender.String(), Amount: "5", ToEmail: "missing@x.com"}
	data, _ := json.Marshal(payload)
	task := asynq.NewTask(TypeTransfer, data)
	if err := qm.handleTransfer(context.Background(), task); err == nil {
		t.Fatalf("esperava erro recipient not found")
	}
}

func TestHandleTransfer_Insufficient(t *testing.T) {
	db := setupLiteDB(t)
	qm := newTestQM(db)
	sender := uuid.New()
	recipient := uuid.New()
	seedUser(t, db, sender, "sender@x.com")
	seedUser(t, db, recipient, "rec@x.com")
	seedBalance(t, db, sender, decimal.NewFromFloat(2))
	payload := TransactionPayload{UserID: sender.String(), Amount: "5", ToEmail: "rec@x.com"}
	data, _ := json.Marshal(payload)
	task := asynq.NewTask(TypeTransfer, data)
	if err := qm.handleTransfer(context.Background(), task); err == nil {
		t.Fatalf("esperava erro insuficiente")
	}
}

func TestHandleTransfer_Success(t *testing.T) {
	db := setupLiteDB(t)
	qm := newTestQM(db)
	sender := uuid.New()
	recipient := uuid.New()
	seedUser(t, db, sender, "sender@x.com")
	seedUser(t, db, recipient, "rec@x.com")
	seedBalance(t, db, sender, decimal.NewFromFloat(20))
	payload := TransactionPayload{UserID: sender.String(), Amount: "5", ToEmail: "rec@x.com"}
	data, _ := json.Marshal(payload)
	task := asynq.NewTask(TypeTransfer, data)
	if err := qm.handleTransfer(context.Background(), task); err != nil {
		t.Fatalf("transfer erro: %v", err)
	}
}
