package services

import (
	"context"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	tests := []struct {
		name         string
		maxFailures  int
		resetTimeout time.Duration
		functions    []struct {
			shouldFail bool
			expectOpen bool
		}
	}{
		{
			name:         "circuit opens after max failures",
			maxFailures:  2,
			resetTimeout: 100 * time.Millisecond,
			functions: []struct {
				shouldFail bool
				expectOpen bool
			}{
				{true, false},  // Fail 1
				{true, true},   // Fail 2, circuit opens
				{false, true},  // Should still be open
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCircuitBreaker("test", tt.maxFailures, tt.resetTimeout)

			for i, fn := range tt.functions {
				var err error
				if fn.shouldFail {
					err = cb.Execute(func() error {
						return &testError{"test failure"}
					})
				} else {
					err = cb.Execute(func() error {
						return nil
					})
				}

				isOpen := cb.GetState() == StateOpen
				if isOpen != fn.expectOpen {
					t.Errorf("test %d: expected open=%v, got %v", i, fn.expectOpen, isOpen)
				}

				if fn.expectOpen && err == nil {
					t.Errorf("test %d: expected error when circuit is open", i)
				}
			}
		})
	}
}

func TestCache(t *testing.T) {
	tests := []struct {
		name    string
		ttl     time.Duration
		maxSize int
		ops     []struct {
			op       string
			key      string
			value    interface{}
			wantHit  bool
			wantVal  interface{}
		}
	}{
		{
			name:    "basic cache operations",
			ttl:     1 * time.Second,
			maxSize: 10,
			ops: []struct {
				op       string
				key      string
				value    interface{}
				wantHit  bool
				wantVal  interface{}
			}{
				{"set", "key1", "value1", false, nil},
				{"get", "key1", nil, true, "value1"},
				{"set", "key2", 42, false, nil},
				{"get", "key2", nil, true, 42},
				{"delete", "key1", nil, false, nil},
				{"get", "key1", nil, false, nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewCache(tt.ttl, tt.maxSize)

			for _, op := range tt.ops {
				switch op.op {
				case "set":
					cache.Set(op.key, op.value)
				case "get":
					val, hit := cache.Get(op.key)
					if hit != op.wantHit {
						t.Errorf("get %s: expected hit=%v, got %v", op.key, op.wantHit, hit)
					}
					if hit && val != op.wantVal {
						t.Errorf("get %s: expected %v, got %v", op.key, op.wantVal, val)
					}
				case "delete":
					cache.Delete(op.key)
				}
			}
		})
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := NewCache(10*time.Millisecond, 10)
	cache.Set("key1", "value1")

	// Should be found
	_, hit := cache.Get("key1")
	if !hit {
		t.Error("expected cache hit immediately after set")
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should not be found
	_, hit = cache.Get("key1")
	if hit {
		t.Error("expected cache miss after expiration")
	}
}

func TestBatchProcessor(t *testing.T) {
	var processed []interface{}

	processor := NewBatchProcessor(2, 100*time.Millisecond, func(batch []interface{}) error {
		processed = append(processed, batch...)
		return nil
	})
	defer processor.Close()

	// Add items that should be batched
	processor.Add("item1")
	processor.Add("item2") // Should trigger batch processing

	time.Sleep(150 * time.Millisecond)

	if len(processed) != 2 {
		t.Errorf("expected 2 items processed, got %d", len(processed))
	}
}

func TestErrorRecoveryManager(t *testing.T) {
	manager := NewErrorRecoveryManager()
	manager.RegisterCircuitBreaker("test-cb", 2, 100*time.Millisecond)
	manager.RegisterRetry("test-retry", 3, 10*time.Millisecond, 100*time.Millisecond)

	// Test successful execution
	err := manager.Execute(context.Background(), "test-retry", func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("expected no error on successful execution, got %v", err)
	}
}

// Benchmarks

func BenchmarkCache(b *testing.B) {
	cache := NewCache(1*time.Minute, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key" + string(rune(i%100))
		cache.Set(key, "value")
		cache.Get(key)
	}
}

func BenchmarkCircuitBreaker(b *testing.B) {
	cb := NewCircuitBreaker("bench", 10, 1*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Execute(func() error {
			return nil
		})
	}
}
