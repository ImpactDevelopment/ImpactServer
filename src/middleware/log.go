package middleware

import "github.com/labstack/echo/v4/middleware"

var Log = middleware.LoggerWithConfig(middleware.LoggerConfig{
	Format: "#${header:X-Request-ID} ${status} ${method} ${host}${uri} latency=${latency} [${latency_human}] error=${error}\n",
})
