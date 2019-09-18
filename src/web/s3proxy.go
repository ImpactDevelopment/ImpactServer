package web

import (
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo"
)

var awsSess = session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))

func S3Proxy(c echo.Context) error {
	file := c.Param("file")

	s3Req, _ := s3.New(awsSess).GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String("impactclient-builds"),
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

	c.Response().Header().Set("Cache-Control", "max-age=300")

	doProxy(c, target)
	return nil
}
