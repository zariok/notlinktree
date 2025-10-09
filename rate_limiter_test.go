package main

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 2)
	ip := "127.0.0.1"

	if !rl.Allow(ip) {
		t.Error("First call should be allowed")
	}
	if !rl.Allow(ip) {
		t.Error("Second call should be allowed")
	}
	if rl.Allow(ip) {
		t.Error("Third call should be rate limited")
	}

	// Wait for window to reset
	time.Sleep(120 * time.Millisecond)
	if !rl.Allow(ip) {
		t.Error("Should allow after window reset")
	}
}

func TestRateLimiter_MultipleIPs(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 1)
	if !rl.Allow("1.1.1.1") {
		t.Error("First call for 1.1.1.1 should be allowed")
	}
	if !rl.Allow("2.2.2.2") {
		t.Error("First call for 2.2.2.2 should be allowed")
	}
	if rl.Allow("1.1.1.1") {
		t.Error("Second call for 1.1.1.1 should be rate limited")
	}
}

func TestRateLimiter_Concurrency(t *testing.T) {
	rl := NewRateLimiter(100*time.Millisecond, 100)
	ip := "127.0.0.1"
	var allowed int32
	done := make(chan struct{})
	for i := 0; i < 200; i++ {
		go func() {
			if rl.Allow(ip) {
				atomic.AddInt32(&allowed, 1)
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 200; i++ {
		<-done
	}
	if allowed > 100 {
		t.Errorf("Allowed more than maxClicks: %d", allowed)
	}
}
