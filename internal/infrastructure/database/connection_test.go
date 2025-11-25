package repositories

import (
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConnectDatabaseFailure(t *testing.T) {
	origOpen := gormOpen
	origSleep := sleepFn
	defer func() { gormOpen = origOpen; sleepFn = origSleep }()
	attempts := 0
	gormOpen = func(dial gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
		attempts++
		return nil, errors.New("boom")
	}
	sleepFn = func(d time.Duration) {}
	db, err := connectDatabaseWithRetry("postgres://invalid", 3, time.Millisecond)
	if err == nil || db != nil {
		t.Fatalf("expected failure, got db=%v err=%v", db, err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts got %d", attempts)
	}
}

func TestConnectDatabaseSuccessAfterRetries(t *testing.T) {
	origOpen := gormOpen
	origSleep := sleepFn
	defer func() { gormOpen = origOpen; sleepFn = origSleep }()
	attempts := 0
	gormOpen = func(dial gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("temporary")
		}
		return gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	}
	sleepFn = func(d time.Duration) {}
	db, err := connectDatabaseWithRetry("postgres://ignored", 5, time.Millisecond)
	if err != nil || db == nil {
		t.Fatalf("expected success on attempt 3 err=%v db=%v", err, db)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts before success got %d", attempts)
	}
}
