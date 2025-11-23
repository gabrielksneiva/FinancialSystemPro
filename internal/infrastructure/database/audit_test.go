package repositories

import (
	"testing"

	"github.com/google/uuid"
)

func TestLogAudit(t *testing.T) {
	if DB == nil {
		t.Skip("Database not initialized")
	}

	auditLog := &AuditLog{
		UserID:     uuid.New(),
		Action:     "TEST_ACTION",
		OldPayload: []byte(`{"old": "value"}`),
		NewPayload: []byte(`{"new": "value"}`),
		IP:         "192.168.1.1",
	}

	err := LogAudit(auditLog)
	if err != nil {
		t.Errorf("LogAudit failed: %v", err)
	}

	if auditLog.ID == 0 {
		t.Error("Expected AuditLog.ID to be set after creation")
	}
}

func TestGetAuditLogs(t *testing.T) {
	if DB == nil {
		t.Skip("Database not initialized")
	}

	userID := uuid.New()

	// Create test records
	for i := 0; i < 5; i++ {
		auditLog := &AuditLog{
			UserID:     userID,
			Action:     "ACTION_" + string(rune(i)),
			OldPayload: []byte(`{}`),
			NewPayload: []byte(`{}`),
			IP:         "192.168.1.1",
		}
		LogAudit(auditLog)
	}

	// Test retrieval
	logs, err := GetAuditLogs(userID, 10, 0)
	if err != nil {
		t.Errorf("GetAuditLogs failed: %v", err)
	}

	if len(logs) == 0 {
		t.Error("Expected to retrieve audit logs")
	}

	// Verify pagination
	if len(logs) > 10 {
		t.Errorf("Expected at most 10 logs, got %d", len(logs))
	}
}

func TestGetAuditLogsByAction(t *testing.T) {
	if DB == nil {
		t.Skip("Database not initialized")
	}

	// Create test record
	auditLog := &AuditLog{
		UserID:     uuid.New(),
		Action:     "UNIQUE_ACTION_TEST",
		OldPayload: []byte(`{}`),
		NewPayload: []byte(`{}`),
		IP:         "192.168.1.1",
	}
	LogAudit(auditLog)

	// Test retrieval by action
	logs, err := GetAuditLogsByAction("UNIQUE_ACTION_TEST", 10, 0)
	if err != nil {
		t.Errorf("GetAuditLogsByAction failed: %v", err)
	}

	if len(logs) == 0 {
		t.Error("Expected to retrieve audit logs by action")
	}

	for _, log := range logs {
		if log.Action != "UNIQUE_ACTION_TEST" {
			t.Errorf("Expected action 'UNIQUE_ACTION_TEST', got '%s'", log.Action)
		}
	}
}

func TestPurgeOldAuditLogs(t *testing.T) {
	if DB == nil {
		t.Skip("Database not initialized")
	}

	// This test would require time manipulation or specific database setup
	// For now, just verify the function runs without error
	err := PurgeOldAuditLogs(7) // Purge logs older than 7 days
	if err != nil {
		t.Errorf("PurgeOldAuditLogs failed: %v", err)
	}
}
