package main

import (
	"github.com/ImpactDevelopment/ImpactServer/src/lib"
	"github.com/ImpactDevelopment/ImpactServer/src/web"
	"net/http"
)

func Router(s lib.HttpServer) {
	s.Match([]string{http.MethodHead, http.MethodGet}, "/changelog", web.Changelog)
	s.Any("/Impact/*", web.ImpactRedirect)
}
