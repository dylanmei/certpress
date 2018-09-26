package main

import (
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func fetch(fileURL string) ([]byte, error) {
	location, err := url.Parse(fileURL)
	if err != nil {
		return []byte{}, err
	}

	var fetcher Fetcher

	if location.Scheme == "s3" {
		aws := session.Must(session.NewSession())
		fetcher = &BucketObjectFetcher{aws}
	} else {
		fetcher = &LocalFileFetcher{}
	}

	return fetcher.Fetch(location)
}

type Fetcher interface {
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
