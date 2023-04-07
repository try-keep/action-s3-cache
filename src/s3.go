package main

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func GetLatestObject(key, bucket string) (string, error) {
	session, err := getSession()
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	response, err := session.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	})
	if err != nil {
		return "", err
	}

	files := response.Contents

	if len(files) < 1 {
		return "", errors.New("failed to find any files matching default key")
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].LastModified.After(*files[j].LastModified)
	})

	return *files[0].Key, nil
}

// PutObject - Upload object to s3 bucket
func PutObject(key, bucket, s3Class string) error {
	session, err := getSession()
	if err != nil {
		log.Fatal(err)
		return err
	}

	file, err := os.Open(key)
	if err != nil {
		return err
	}
	defer file.Close()
	fileSize, err := file.Stat()
	if err != nil {
		return err
	}

	i := &s3.PutObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(key),
		Body:         file,
		StorageClass: types.StorageClass(s3Class),
	}

	start := time.Now()
	log.Printf("Uploading %v worth of cache", getReadableBytes(fileSize.Size()))
	_, err = session.PutObject(context.TODO(), i)
	if err == nil {
		elapsed := time.Since(start)
		log.Printf("Cache saved successfully in %s!", elapsed)
	}

	return err
}

// GetObject - Get object from s3 bucket
func GetObject(key, bucket string) error {
	start := time.Now()
	session, err := getSession()
	if err != nil {
		log.Fatal(err)
		return err
	}

	i := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	object, err := session.GetObject(context.TODO(), i)
	if err != nil {
		return err
	}
	outFile, err := os.Create(key)
	if err != nil {
		return err
	}

	defer outFile.Close()
	_, err = io.Copy(outFile, object.Body)
	elapsed := time.Since(start)
	log.Printf("%s worth of cache successfully downloaded in %s", getReadableBytes(object.ContentLength), elapsed)
	return err
}

// DeleteObject - Delete object from s3 bucket
func DeleteObject(key, bucket string) error {
	session, err := getSession()
	if err != nil {
		log.Fatal(err)
		return err
	}

	i := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err = session.DeleteObject(context.TODO(), i)
	if err == nil {
		log.Print("Cache purged successfully")
	}

	return err
}

// ObjectExists - Verify if object exists in s3
func ObjectExists(key, bucket string) (bool, error) {
	session, err := getSession()
	if err != nil {
		log.Fatal(err)
		return false, err
	}

	i := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if _, err = session.HeadObject(context.TODO(), i); err != nil {
		return false, nil
	}
	return true, nil
}

func getSession() (*s3.Client, error) {
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")
	if sessionToken == "" {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return nil, err
		}
		return s3.NewFromConfig(cfg), nil
	}
	session := s3.NewFromConfig(aws.Config{
		Region: os.Getenv("AWS_REGION"),
		Credentials: credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			os.Getenv("AWS_SESSION_TOKEN"),
		),
	})
	return session, nil
}
