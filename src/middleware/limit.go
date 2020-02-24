package middleware

import (
	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"
)

func Limit(duration time.Duration, bursts int) echo.MiddlewareFunc {
	limiter := &rateLimiter{
		keys:   make(map[string]*rate.Limiter),
		mutex:  &sync.RWMutex{},
		rate:   rate.Limit(float64(time.Second) / float64(duration)),
		bursts: bursts,
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// key is either user id or user ip depending on if we are logged in
			var key string
			if user, ok := c.Get("user").(*users.User); ok && user != nil {
				key = user.ID.String()
			} else {
				// Get the user's ip; can't just use the request address since we are behind proxies
				key = util.RealIPBestGuess(c)
			}

			// Check we aren't limited
			if !limiter.get(key).Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests)
			}

			return next(c)
		}
	}
}

type rateLimiter struct {
	keys   map[string]*rate.Limiter
	mutex  *sync.RWMutex
	rate   rate.Limit
	bursts int
}

// add creates a new rate limiter if it does not exist already and adds it to the keys map
// calling get() will be cheaper in most cases since it only needs a read lock
func (lim *rateLimiter) add(key string) *rate.Limiter {
	// shit getting real, let's get a full read-write lock!
	lim.mutex.Lock()
	defer lim.mutex.Unlock()

	// Double check nothing was added since last calling RUnlock()
	limiter, exists := lim.keys[key]
	if !exists {
		limiter := rate.NewLimiter(lim.rate, lim.bursts)
		lim.keys[key] = limiter
	}

	return limiter
}

// get returns the rate limiter for the provided key if it exists.
// Otherwise calls add to add key to the map
func (lim *rateLimiter) get(key string) *rate.Limiter {
	// we only need a read lock, for now
	lim.mutex.RLock()
	limiter, exists := lim.keys[key]
	lim.mutex.RUnlock()

	if !exists {
		limiter = lim.add(key)
	}

	return limiter
}
