package middleware

import "github.com/labstack/echo/v4/middleware"

var Log = middleware.LoggerWithConfig(middleware.LoggerConfig{
	Format: "${status} ${method} ${uri} latency=${latency_human} error=${error}\n",
})
