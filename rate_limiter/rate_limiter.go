package rate_limiter

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	noMoreRequests = 0
)

type RateLimiter struct {
	MaxRequests int
	RefreshDuration    time.Duration
	Requests    map[string]int
	IncrementAmount int
	sync.Mutex
}

func NewRateLimiter(
	maxRequests int, 
	refreshDuration time.Duration,
	incrementAmount int) *RateLimiter {

	message := fmt.Sprintf("RateLimiter::NewRateLimiter - New RateLimiter created with maxRequests: %d, refreshDuration: %v, incrementAmount: %d", maxRequests, refreshDuration, incrementAmount)
	log.Default().Println(message)

	return &RateLimiter{
		MaxRequests: maxRequests,
		IncrementAmount: incrementAmount,
		RefreshDuration:    refreshDuration,
		Requests:    make(map[string]int),
	}
}

func (rl *RateLimiter) Refresh() {
    ticker := time.NewTicker(rl.RefreshDuration)
	for range ticker.C {
		log.Default().Println("RateLimiter::Refresh - Refreshing request limits")
		for ip, numRequests := range rl.Requests {
			rl.Lock()
			rl.Requests[ip] = numRequests + rl.IncrementAmount

			if rl.Requests[ip] > rl.MaxRequests {
				rl.Requests[ip] = rl.MaxRequests
			}

			message := fmt.Sprintf("RateLimiter::Refresh - IP %s has %d requests remaining after refresh", ip, rl.Requests[ip])
			log.Default().Println(message)
			rl.Unlock()
		}	
	}	
}

func (rl *RateLimiter) RateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rl.Lock()
		defer rl.Unlock()

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Default().Println("RateLimiter::Limit - Error getting remote address: ", err)
			http.Error(w, "Error getting remote address", http.StatusInternalServerError)
			return
		}

		log.Default().Println("RateLimiter::Limit - incoming request from IP: ", ip)

		var isNewIp bool
		_, ok := rl.Requests[ip]
		if !ok {
			isNewIp = true
			message := fmt.Sprintf("RateLimiter::Limit - IP %s not found in requests map - adding %d requests", ip, rl.MaxRequests)
			log.Default().Println(message)
			rl.Requests[ip] = rl.MaxRequests
		}

		if rl.Requests[ip] <= noMoreRequests {
			message := fmt.Sprintf("RateLimiter::Limit - IP %s has no more requests - will wait for refresh", ip)
			log.Default().Println(message)
			w.Header().Set("X-Retry-After", rl.RefreshDuration.String())
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		if !isNewIp {
			rl.Requests[ip]--
		}

		message := fmt.Sprintf("RateLimiter::Limit - IP %s has %d requests remaining", ip, rl.Requests[ip])
		log.Default().Println(message)
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rl.Requests[ip]))
		next.ServeHTTP(w, r)
	}
}
