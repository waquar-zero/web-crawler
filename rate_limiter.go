package main

import (
	"context"
	"net/url"
	"sync"
	"time"
)

type HostLimiter struct {
	mu       sync.Mutex
	nextTime map[string]time.Time
	delay    time.Duration
}

func NewHostLimiter(delay time.Duration) *HostLimiter {
	return &HostLimiter{
		nextTime: make(map[string]time.Time),
		delay:    delay,
	}
}

func (h *HostLimiter) Wait(ctx context.Context, rawURL string) error {
	if err := ctx.Err(); err != nil {
    return err
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	host := parsed.Scheme + "://" + parsed.Host

	h.mu.Lock()

	now := time.Now()
	scheduled := h.nextTime[host]

	if scheduled.Before(now) {
		scheduled = now
	}

	h.nextTime[host] = scheduled.Add(h.delay)

	h.mu.Unlock()

	wait := time.Until(scheduled)

	if wait <= 0 {
		return ctx.Err()
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-timer.C:
		return ctx.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}
