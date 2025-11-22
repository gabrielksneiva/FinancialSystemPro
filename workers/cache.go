package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// CacheManager gerencia cache em Redis para operações críticas
type CacheManager struct {
	client  *redis.Client
	ttl     time.Duration
	logger  *zap.Logger
	breaker *CircuitBreaker
}

// NewCacheManager cria um gerenciador de cache
func NewCacheManager(redisURL string, logger *zap.Logger, breaker *CircuitBreaker) *CacheManager {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Error("failed to parse redis url for cache", zap.Error(err))
		return nil
	}

	client := redis.NewClient(opt)

	return &CacheManager{
		client:  client,
		ttl:     5 * time.Minute, // Cache por 5 minutos
		logger:  logger,
		breaker: breaker,
	}
}

// GetBalance retorna saldo em cache (ou vazio se expirado/miss)
func (cm *CacheManager) GetBalance(ctx context.Context, userID string) (decimal.Decimal, bool) {
	key := fmt.Sprintf("balance:%s", userID)

	err := cm.breaker.Call(func() error {
		val, err := cm.client.Get(ctx, key).Result()
		if err == redis.Nil {
			return nil // Cache miss é ok
		}
		return err
	})

	if err != nil {
		cm.logger.Debug("cache get failed", zap.String("key", key), zap.Error(err))
		return decimal.Zero, false
	}

	val, err := cm.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return decimal.Zero, false // Cache miss
	}
	if err != nil {
		return decimal.Zero, false
	}

	balance, err := decimal.NewFromString(val)
	if err != nil {
		cm.logger.Warn("failed to parse cached balance", zap.String("value", val), zap.Error(err))
		return decimal.Zero, false
	}

	return balance, true
}

// SetBalance armazena saldo em cache
func (cm *CacheManager) SetBalance(ctx context.Context, userID string, balance decimal.Decimal) error {
	key := fmt.Sprintf("balance:%s", userID)

	return cm.breaker.Call(func() error {
		return cm.client.Set(ctx, key, balance.String(), cm.ttl).Err()
	})
}

// InvalidateBalance remove saldo do cache (após transação)
func (cm *CacheManager) InvalidateBalance(ctx context.Context, userID string) error {
	key := fmt.Sprintf("balance:%s", userID)

	return cm.breaker.Call(func() error {
		return cm.client.Del(ctx, key).Err()
	})
}

// CacheTaskResult armazena resultado de tarefa para retry
func (cm *CacheManager) CacheTaskResult(ctx context.Context, taskID string, result interface{}, ttl time.Duration) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return cm.breaker.Call(func() error {
		return cm.client.Set(ctx, fmt.Sprintf("task-result:%s", taskID), string(data), ttl).Err()
	})
}

// GetTaskResult recupera resultado em cache
func (cm *CacheManager) GetTaskResult(ctx context.Context, taskID string) (interface{}, error) {
	key := fmt.Sprintf("task-result:%s", taskID)
	var result interface{}

	err := cm.breaker.Call(func() error {
		val, err := cm.client.Get(ctx, key).Result()
		if err == redis.Nil {
			return nil
		}
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(val), &result)
	})

	return result, err
}

// SetDeduplication previne processamento duplicado de tarefas
func (cm *CacheManager) SetDeduplication(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	dedupKey := fmt.Sprintf("dedup:%s", key)
	var success bool

	err := cm.breaker.Call(func() error {
		// SetNX = Set if Not eXists
		ok, err := cm.client.SetNX(ctx, dedupKey, value, ttl).Result()
		success = ok
		return err
	})

	return success, err
}

// GetDeduplication verifica se tarefa já foi processada
func (cm *CacheManager) GetDeduplication(ctx context.Context, key string) (string, error) {
	dedupKey := fmt.Sprintf("dedup:%s", key)
	var result string

	err := cm.breaker.Call(func() error {
		val, err := cm.client.Get(ctx, dedupKey).Result()
		if err == redis.Nil {
			return nil
		}
		result = val
		return err
	})

	return result, err
}

// IsHealthy verifica se cache está funcionando
func (cm *CacheManager) IsHealthy(ctx context.Context) bool {
	return cm.breaker.IsHealthy()
}

// Close fecha conexão
func (cm *CacheManager) Close() error {
	return cm.client.Close()
}
