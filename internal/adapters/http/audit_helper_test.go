package http

import (
	"net/http/httptest"
	"testing"

	repositories "financial-system-pro/internal/infrastructure/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuditDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	repositories.DB = db
	if err := db.AutoMigrate(&repositories.AuditLog{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
}

func acquireCtx(t *testing.T) *fiber.Ctx {
	app := fiber.New()
	var captured *fiber.Ctx
	app.Get("/", func(c *fiber.Ctx) error { captured = c; return c.SendStatus(200) })
	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	resp, err := app.Test(req, -1)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		t.Fatalf("failed to get ctx: %v", err)
	}
	return captured
}

func TestAuditLogHelper_AllActions(t *testing.T) {
	setupAuditDB(t)
	ctx := acquireCtx(t)
	helper := NewAuditLogHelper(zap.NewNop())
	uid := uuid.New()

	// Call each with success and failure to exercise both branches
	helper.LogLoginAttempt(uid, "user@example.com", true, ctx)
	helper.LogLoginAttempt(uid, "user@example.com", false, ctx)
	helper.LogDeposit(uid, "100.00", true, ctx)
	helper.LogDeposit(uid, "100.00", false, ctx)
	helper.LogWithdraw(uid, "50.00", true, ctx)
	helper.LogWithdraw(uid, "50.00", false, ctx)
	helper.LogTransfer(uid, uuid.New().String(), "25.00", true, ctx)
	helper.LogTransfer(uid, uuid.New().String(), "25.00", false, ctx)
	helper.LogTronTransaction(uid, "TRANSFER", "TADDR123", "10", true, ctx)
	helper.LogTronTransaction(uid, "TRANSFER", "TADDR123", "10", false, ctx)
	helper.LogUserCreation(uid, "new@example.com", true, ctx)
	helper.LogUserCreation(uid, "new@example.com", false, ctx)

	// Count records
	var count int64
	if err := repositories.DB.Model(&repositories.AuditLog{}).Count(&count).Error; err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 12 {
		t.Fatalf("expected 12 audit records got %d", count)
	}
}
