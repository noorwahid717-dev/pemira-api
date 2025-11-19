package middleware

import (
	"net/http"
	"sync"
	"time"

	"pemira-api/internal/http/response"
)

type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int
	burst    int
}

type visitor struct {
	lastSeen time.Time
	tokens   int
}

func NewRateLimiter(requestsPerMinute, burst int) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     requestsPerMinute,
		burst:    burst,
	}

	// Cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

func (rl *rateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			lastSeen: time.Now(),
			tokens:   rl.burst,
		}
		rl.visitors[ip] = v
	}

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(v.lastSeen)
	tokensToAdd := int(elapsed.Seconds() * float64(rl.rate) / 60.0)
	
	v.tokens += tokensToAdd
	if v.tokens > rl.burst {
		v.tokens = rl.burst
	}
	v.lastSeen = now

	return v
}

func (rl *rateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		
		v := rl.getVisitor(ip)
		
		if v.tokens <= 0 {
			response.Error(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests", nil)
			return
		}

		v.tokens--
		next.ServeHTTP(w, r)
	})
}
