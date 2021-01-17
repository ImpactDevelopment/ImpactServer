package s3proxy

import (
	mid "github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"net/http"
	"net/url"
	"time"
	"os"

	"github.com/ImpactDevelopment/ImpactServer/src/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo/v4"
)

var AWSSession = session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))

func Server() (e *echo.Echo) {
	e = echo.New()

	e.Use(mid.Log)

	e.Match([]string{http.MethodHead, http.MethodGet}, "/*", proxyHandler("", "impactclient-files"))
	e.Match([]string{http.MethodHead, http.MethodGet}, "/test_alternate/*", proxyHandler("/test_alternate", os.Getenv("ALT_BUCKET")), mid.AuthGetParam(), mid.NoCache())

	return
}

func proxyHandler(base string, bucket string) func (c echo.Context) error {
	return func (c echo.Context) error {
		file := c.Request().URL.Path[len(base):]

		s3Req, _ := s3.New(AWSSession).GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
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
}
