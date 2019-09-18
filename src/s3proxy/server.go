package web

import (
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/labstack/echo/middleware"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo"
)

var awsSess = session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))

func Server() (e *echo.Echo) {
	e = echo.New()

	e.Match([]string{http.MethodHead, http.MethodGet}, "/*", proxyHandler)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	return
}

func proxyHandler(c echo.Context) error {
	file := c.Request().URL.Path

	s3Req, _ := s3.New(awsSess).GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String("impactclient-static"),
		Key:    aws.String(file),
	})

	s3PresignedURL, err := s3Req.Presign(1 * time.Minute)
	if err != nil {
		return err
	}

	target, err := url.Parse(s3PresignedURL)
	if err != nil {
		return err
	}

	util.Proxy(c, target)
	return nil
}
