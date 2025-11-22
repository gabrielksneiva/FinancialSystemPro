package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// TronGRPCClient representa um cliente gRPC para Tron
type TronGRPCClient struct {
	endpoint    string
	conn        *grpc.ClientConn
	mu          sync.Mutex
	isConnected bool
}

// NewTronGRPCClient cria um novo cliente gRPC para Tron
func NewTronGRPCClient(endpoint string) (*TronGRPCClient, error) {
	client := &TronGRPCClient{
		endpoint: endpoint,
	}

	// Configurar options de keepalive para melhor performance
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             3 * time.Second,
		PermitWithoutStream: true,
	}

	// Configurar credenciais TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}
	creds := credentials.NewTLS(tlsConfig)

	// Conectar ao servidor gRPC
	conn, err := grpc.DialContext(
		context.Background(),
		endpoint,
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(50*1024*1024),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao gRPC Tron: %w", err)
	}

	client.conn = conn
	client.isConnected = true

	return client, nil
}

// IsConnected verifica se está conectado
func (c *TronGRPCClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return false
	}

	state := c.conn.GetState()
	return state.String() == "READY"
}

// Reconnect reconecta ao servidor gRPC
func (c *TronGRPCClient) Reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
	}

	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             3 * time.Second,
		PermitWithoutStream: true,
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}
	creds := credentials.NewTLS(tlsConfig)

	conn, err := grpc.DialContext(
		context.Background(),
		c.endpoint,
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(50*1024*1024),
		),
	)
	if err != nil {
		c.isConnected = false
		return fmt.Errorf("erro ao reconectar ao gRPC Tron: %w", err)
	}

	c.conn = conn
	c.isConnected = true
	return nil
}

// GetConnection retorna a conexão gRPC
func (c *TronGRPCClient) GetConnection() *grpc.ClientConn {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn
}

// Close fecha a conexão gRPC
func (c *TronGRPCClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.isConnected = false
		return err
	}

	return nil
}

// HealthCheck realiza um health check na conexão
func (c *TronGRPCClient) HealthCheck(ctx context.Context) error {
	if !c.IsConnected() {
		return fmt.Errorf("cliente gRPC não está conectado")
	}

	// Timeout para health check
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn := c.GetConnection()
	if conn == nil {
		return fmt.Errorf("conexão gRPC é nula")
	}

	// Tentar obter o estado da conexão
	if state := conn.GetState(); state.String() != "READY" {
		return fmt.Errorf("conexão gRPC em estado: %s", state.String())
	}

	return nil
}
