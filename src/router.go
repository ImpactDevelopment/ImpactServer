package main

import (
	"github.com/ImpactDevelopment/ImpactServer/src/lib"
	"github.com/ImpactDevelopment/ImpactServer/src/web"
)

func Router(s lib.HttpServer) {
	s.Match([]string{"HEAD", "GET"}, "/changelog", web.Changelog)
	s.Any("/Impact/*", web.ImpactRedirect)
}
