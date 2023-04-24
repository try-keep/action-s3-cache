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
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func GetLatestObject(key, bucket string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	session := s3.NewFromConfig(cfg)
	if err != nil {
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
	cfg, err := config.LoadDefaultConfig(context.TODO())
	session := s3.NewFromConfig(cfg)
	if err != nil {
		return err
	}

	uploader := manager.NewUploader(session, func(u *manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // 10 MiB
		u.Concurrency = 4
		u.MaxUploadParts = 50
	})

	file, err := os.Open(key)
	if err != nil {
		return err
	}
	defer file.Close()
	fileSize, err := file.Stat()
	if err != nil {
		return err
	}

	start := time.Now()
	log.Printf("Uploading %v worth of cache", getReadableBytes(fileSize.Size()))
	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(key),
		Body:         file,
		StorageClass: types.StorageClass(s3Class),
	})
	if err == nil {
		elapsed := time.Since(start)
		log.Printf("Cache saved successfully in %s!", elapsed)
	}

	return err
}

// GetObject - Get object from s3 bucket
func GetObject(key, bucket string) error {
	start := time.Now()
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}

	session := s3.NewFromConfig(cfg)

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
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	session := s3.NewFromConfig(cfg)

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
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return false, err
	}
	session := s3.NewFromConfig(cfg)

	i := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if _, err = session.HeadObject(context.TODO(), i); err != nil {
		return false, nil
	}
	return true, nil
}
