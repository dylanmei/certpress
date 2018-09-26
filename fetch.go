package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func fetch(fileURL string) ([]byte, error) {
	location, err := url.Parse(fileURL)
	if err != nil {
		return []byte{}, err
	}

	var fetcher FileFetcher

	if location.Scheme == "s3" {
		fetcher = &BucketObjectFetcher{awsSession()}
	} else {
		fetcher = &LocalFileFetcher{}
	}

	return fetcher.Fetch(location)
}

func awsSession() *session.Session {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		sess := session.Must(session.NewSession())
		ec2m := ec2metadata.New(sess)
		if ec2m.Available() {
			region, _ = ec2m.Region()
		}
	}
	if region == "" {
		return session.Must(session.NewSession())
	}

	return session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
}

type FileFetcher interface {
	Fetch(location *url.URL) ([]byte, error)
}

type LocalFileFetcher struct {
}

func (fetcher *LocalFileFetcher) Fetch(location *url.URL) ([]byte, error) {
	return ioutil.ReadFile(location.Path)
}

type BucketObjectFetcher struct {
	config client.ConfigProvider
}

func (fetcher *BucketObjectFetcher) Fetch(location *url.URL) ([]byte, error) {
	//fmt.Printf("BUCKET: %s, PATH: %s\n", location.Host, location.Path)
	downloader := s3manager.NewDownloader(fetcher.config)
	var err error
	b := aws.NewWriteAtBuffer([]byte{})

	_, err = downloader.Download(b, &s3.GetObjectInput{
		Bucket: aws.String(location.Host),
		Key:    aws.String(location.Path),
	})

	if err != nil {
		return []byte{}, fmt.Errorf("Failed to download S3 object, %v", err)
	}

	return b.Bytes(), nil
}
