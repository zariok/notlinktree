package main

import (
	"sync"
	"time"
)

type RateLimiter struct {
	clicks     map[string]int       // IP -> click count
	lastReset  map[string]time.Time // IP -> last reset time
	mu         sync.RWMutex
	windowSize time.Duration
	maxClicks  int
}

func NewRateLimiter(windowSize time.Duration, maxClicks int) *RateLimiter {
	return &RateLimiter{
		clicks:     make(map[string]int),
		lastReset:  make(map[string]time.Time),
		windowSize: windowSize,
		maxClicks:  maxClicks,
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	lastReset, exists := rl.lastReset[ip]

	// Reset counter if window has passed
	if !exists || now.Sub(lastReset) > rl.windowSize {
		rl.clicks[ip] = 0
		rl.lastReset[ip] = now
	}

	// Check if under limit
	if rl.clicks[ip] >= rl.maxClicks {
		return false
	}

	rl.clicks[ip]++
	return true
}
