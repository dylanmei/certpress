package certpress

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

func fetchBytes(fileURL string) ([]byte, error) {
	location, err := url.Parse(fileURL)
	if err != nil {
		return []byte{}, err
	}

	if location.Scheme == "s3" {
		return fetchBytesFromBucketObject(awsSession(), location.Host, location.Path)
	} else if location.Scheme == "awssm" {
		return fetchBytesFromSecretsManager(awsSession(), location.Host, strings.TrimPrefix(location.Path, "/"))
	}

	return fetchBytesFromFilesystem(location.Path)
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

func fetchBytesFromFilesystem(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

func fetchBytesFromBucketObject(config client.ConfigProvider, bucketName, objectName string) ([]byte, error) {
	downloader := s3manager.NewDownloader(config)
	var err error
	b := aws.NewWriteAtBuffer([]byte{})

	_, err = downloader.Download(b, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to download S3 object from %s, %v", bucketName, err)
	}

	return b.Bytes(), nil
}

func fetchBytesFromSecretsManager(config client.ConfigProvider, secretName, secretKey string) ([]byte, error) {
	manager := secretsmanager.New(config)
	input := secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := manager.GetSecretValue(&input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			return nil, fmt.Errorf("Failed to get secret %s value: %v", secretName, aerr.Error())
		}

		return nil, err
	}

	if result.SecretString != nil {
		secretString := *result.SecretString
		secretBytes := []byte(secretString)

		// If the secret spec contains a path component,
		// it is a key into dictionary of key/values.
		if secretKey != "" {
			secretsMap := make(map[string]string)
			err := json.Unmarshal(secretBytes, &secretsMap)
			if err != nil {
				return nil, err
			}

			if _, ok := secretsMap[secretKey]; !ok {
				return nil, fmt.Errorf("Key %s is not a component of secret %s", secretKey, secretName)
			}

			// AWS Secrets Manager may encode newlines characters
			// as spaces when the secret is a dictionary.
			certificate := pemFixupWhitespace(secretsMap[secretKey])
			secretBytes = []byte(certificate)
		}

		return secretBytes, nil
	} else {
		return nil, fmt.Errorf("Decoding binary secrets is not implemented!")
	}
}

func pemFixupWhitespace(text string) string {
	if len(text) == 0 {
		return text
	}

	layoutRegex := regexp.MustCompile(`-{5}[\s\w]+-{5}`)
	layoutArray := layoutRegex.Split(text, 3)
	labelsArray := layoutRegex.FindAllString(text, 2)

	if len(layoutArray) != 3 {
		//panic("Certificate layout does not contain enough sections")
		return text
	}

	if len(labelsArray) != 2 {
		//panic("Certificate layout does not contain enough labels")
		return text
	}

	encodedData := strings.ReplaceAll(strings.Trim(layoutArray[1], " \n"), " ", "\n")
	certificate := fmt.Sprintf("%s\n%s\n%s", labelsArray[0], encodedData, labelsArray[1])
	return certificate
}
