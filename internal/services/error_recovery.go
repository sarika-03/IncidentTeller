package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "closed"
	StateOpen     CircuitBreakerState = "open"
	StateHalfOpen CircuitBreakerState = "half-open"
)

// CircuitBreaker implements the circuit breaker pattern for fault tolerance
type CircuitBreaker struct {
	name          string
	maxFailures   int
	resetTimeout  time.Duration
	state         CircuitBreakerState
	failures      int
	successCount  int
	lastFailTime  time.Time
	mu            sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:         name,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
		failures:     0,
	}
}

// Execute runs a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if circuit should reset
	if cb.state == StateOpen {
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			cb.failures = 0
		} else {
			return fmt.Errorf("circuit breaker %s is open", cb.name)
		}
	}

	// Execute the function
	err := fn()

	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()

		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}

		return err
	}

	// Success
	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= 2 {
			cb.state = StateClosed
			cb.failures = 0
		}
	} else if cb.state == StateClosed {
		cb.failures = 0
	}

	return nil
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Retry implements the retry pattern with exponential backoff
type Retry struct {
	maxAttempts int
	initialDelay time.Duration
	maxDelay     time.Duration
	multiplier  float64
}

// NewRetry creates a new retry configuration
func NewRetry(maxAttempts int, initialDelay time.Duration, maxDelay time.Duration) *Retry {
	return &Retry{
		maxAttempts:  maxAttempts,
		initialDelay: initialDelay,
		maxDelay:     maxDelay,
		multiplier:   2.0,
	}
}

// Execute runs a function with retry logic
func (r *Retry) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	var lastErr error
	delay := r.initialDelay

	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		// Check context before attempting
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Try to execute
		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry if it's the last attempt
		if attempt == r.maxAttempts {
			break
		}

		// Wait with backoff
		select {
		case <-time.After(delay):
			// Calculate next delay
			delay = time.Duration(float64(delay) * r.multiplier)
			if delay > r.maxDelay {
				delay = r.maxDelay
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", r.maxAttempts, lastErr)
}

// BulkheadExecutor limits concurrent executions
type BulkheadExecutor struct {
	name      string
	semaphore chan struct{}
	mu        sync.RWMutex
}

// NewBulkheadExecutor creates a new bulkhead executor
func NewBulkheadExecutor(name string, maxConcurrent int) *BulkheadExecutor {
	return &BulkheadExecutor{
		name:      name,
		semaphore: make(chan struct{}, maxConcurrent),
	}
}

// Execute runs a function with concurrency limits
func (be *BulkheadExecutor) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	select {
	case be.semaphore <- struct{}{}:
		defer func() { <-be.semaphore }()
		return fn(ctx)
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("bulkhead %s at capacity", be.name)
	}
}

// TimeoutExecutor adds timeout protection
type TimeoutExecutor struct {
	timeout time.Duration
}

// NewTimeoutExecutor creates a new timeout executor
func NewTimeoutExecutor(timeout time.Duration) *TimeoutExecutor {
	return &TimeoutExecutor{
		timeout: timeout,
	}
}

// Execute runs a function with timeout
func (te *TimeoutExecutor) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	newCtx, cancel := context.WithTimeout(ctx, te.timeout)
	defer cancel()

	return fn(newCtx)
}

// FallbackHandler provides fallback responses when primary fails
type FallbackHandler struct {
	primaryFn   func(context.Context) (interface{}, error)
	fallbackFn  func(context.Context) (interface{}, error)
}

// NewFallbackHandler creates a new fallback handler
func NewFallbackHandler(
	primary func(context.Context) (interface{}, error),
	fallback func(context.Context) (interface{}, error),
) *FallbackHandler {
	return &FallbackHandler{
		primaryFn:  primary,
		fallbackFn: fallback,
	}
}

// Execute tries primary, then fallback on failure
func (fh *FallbackHandler) Execute(ctx context.Context) (interface{}, error) {
	result, err := fh.primaryFn(ctx)
	if err == nil {
		return result, nil
	}

	// Try fallback
	return fh.fallbackFn(ctx)
}

// ErrorRecoveryManager manages error recovery strategies
type ErrorRecoveryManager struct {
	circuitBreakers map[string]*CircuitBreaker
	retries         map[string]*Retry
	bulkheads       map[string]*BulkheadExecutor
	mu              sync.RWMutex
}

// NewErrorRecoveryManager creates a new error recovery manager
func NewErrorRecoveryManager() *ErrorRecoveryManager {
	return &ErrorRecoveryManager{
		circuitBreakers: make(map[string]*CircuitBreaker),
		retries:         make(map[string]*Retry),
		bulkheads:       make(map[string]*BulkheadExecutor),
	}
}

// RegisterCircuitBreaker registers a circuit breaker
func (erm *ErrorRecoveryManager) RegisterCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) {
	erm.mu.Lock()
	defer erm.mu.Unlock()
	erm.circuitBreakers[name] = NewCircuitBreaker(name, maxFailures, resetTimeout)
}

// RegisterRetry registers a retry strategy
func (erm *ErrorRecoveryManager) RegisterRetry(name string, maxAttempts int, initialDelay, maxDelay time.Duration) {
	erm.mu.Lock()
	defer erm.mu.Unlock()
	erm.retries[name] = NewRetry(maxAttempts, initialDelay, maxDelay)
}

// RegisterBulkhead registers a bulkhead executor
func (erm *ErrorRecoveryManager) RegisterBulkhead(name string, maxConcurrent int) {
	erm.mu.Lock()
	defer erm.mu.Unlock()
	erm.bulkheads[name] = NewBulkheadExecutor(name, maxConcurrent)
}

// Execute runs a function with full error recovery
func (erm *ErrorRecoveryManager) Execute(ctx context.Context, name string, fn func(context.Context) error) error {
	erm.mu.RLock()
	cb := erm.circuitBreakers[name]
	retry := erm.retries[name]
	bulkhead := erm.bulkheads[name]
	erm.mu.RUnlock()

	// Check circuit breaker
	if cb != nil {
		if cb.GetState() == StateOpen {
			return errors.New("circuit breaker is open")
		}
	}

	// Use bulkhead if available
	executeFunc := fn
	if bulkhead != nil {
		executeFunc = func(ctx context.Context) error {
			return bulkhead.Execute(ctx, fn)
		}
	}

	// Use retry if available
	if retry != nil {
		err := retry.Execute(ctx, executeFunc)
		if cb != nil {
			cb.Execute(func() error { return err })
		}
		return err
	}

	// Direct execution
	err := executeFunc(ctx)
	if cb != nil {
		cb.Execute(func() error { return err })
	}
	return err
}
