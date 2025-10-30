package middleware

import (
	"net/http"
	"sync"
	"time"
)

type IPRateLimiter struct {
	ips      map[string]*rateLimiter
	mu       sync.RWMutex
	rate     time.Duration
	burst    int
	cleanup  *time.Ticker
	stopChan chan struct{}
}

type rateLimiter struct {
	tokens chan struct{}
	last   time.Time
}

func NewIPRateLimiter(requests int, per time.Duration, burst int) *IPRateLimiter {
	rl := &IPRateLimiter{
		ips:      make(map[string]*rateLimiter),
		rate:     per / time.Duration(requests),
		burst:    burst,
		cleanup:  time.NewTicker(time.Minute),
		stopChan: make(chan struct{}),
	}

	go rl.cleanupOldEntries()
	return rl
}

func (rl *IPRateLimiter) getLimiter(ip string) *rateLimiter {
	rl.mu.RLock()
	limiter, exists := rl.ips[ip]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		limiter = &rateLimiter{
			tokens: make(chan struct{}, rl.burst),
			last:   time.Now(),
		}

		for i := 0; i < rl.burst; i++ {
			limiter.tokens <- struct{}{}
		}
		rl.ips[ip] = limiter
		rl.mu.Unlock()
	}

	return limiter
}

func (rl *IPRateLimiter) cleanupOldEntries() {
	for {
		select {
		case <-rl.cleanup.C:
			rl.mu.Lock()
			for ip, limiter := range rl.ips {
				if time.Since(limiter.last) > 1*time.Minute {
					delete(rl.ips, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopChan:
			rl.cleanup.Stop()
			return
		}
	}
}

func (rl *IPRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getRealIP(r)
		limiter := rl.getLimiter(ip)

		select {
		case <-limiter.tokens:
			go rl.refillToken(limiter)
			next.ServeHTTP(w, r)
		default:
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
		}
	})
}

func (rl *IPRateLimiter) refillToken(limiter *rateLimiter) {
	time.Sleep(rl.rate)
	select {
	case limiter.tokens <- struct{}{}:
		limiter.last = time.Now()
	default:
	}
}

func (rl *IPRateLimiter) Stop() {
	close(rl.stopChan)
}

func getRealIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
