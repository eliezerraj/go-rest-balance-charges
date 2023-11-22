package circuitbreaker

import (
	"github.com/sony/gobreaker"
	"time"
    "github.com/go-rest-balance-charges/internal/erro"
)

func CircuitBreakerConfig() *gobreaker.CircuitBreaker {
    settings := gobreaker.Settings{
                                        Name:    "server-circuit-breaker",
                                        Timeout: 5 * time.Second,
                                        Interval: 10 * time.Second,
                                        IsSuccessful: func(err error) bool {
                                            if (err == erro.ErrNotFound) || (err == nil) {
                                                return true
                                            } 
                                            return false
                                        },
                                        ReadyToTrip: func(counts gobreaker.Counts) bool {
                                            return counts.TotalFailures >= 3
                                        },
    }
    return gobreaker.NewCircuitBreaker(settings)
}