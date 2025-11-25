package database

import "testing"

// TestNewPostgresConnection_InvalidDSN cobre caminho de erro ao conectar.
func TestNewPostgresConnection_InvalidDSN(t *testing.T) {
	conn, err := NewPostgresConnection("postgres://invalid-host:5432/bad?sslmode=disable")
	if err == nil || conn != nil {
		t.Fatalf("esperava erro e conexão nil, got err=%v conn=%v", err, conn)
	}
}
