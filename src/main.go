package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	action := Action{
		Action:     os.Getenv("ACTION"),
		Bucket:     os.Getenv("BUCKET"),
		S3Class:    os.Getenv("S3_CLASS"),
		Key:        fmt.Sprintf("%s.tar.gz", os.Getenv("KEY")),
		DefaultKey: os.Getenv("DEFAULT_KEY"),
		Artifacts:  strings.Split(strings.TrimSpace(os.Getenv("ARTIFACTS")), "\n"),
	}

	switch act := action.Action; act {
	case PutAction:
		if len(action.Artifacts[0]) <= 0 {
			log.Fatal("No artifacts patterns provided")
		}

		shouldSkip, err := ObjectExists(action.Key, action.Bucket)
		if err != nil {
			log.Fatal(err)
		}
		if shouldSkip {
			log.Printf("Cache hit! Skipping cache upload!")
			return
		} else {
			log.Printf("Cache miss")
		}

		if err := Zip(action.Key, action.Artifacts); err != nil {
			log.Fatal(err)
		}

		if err := PutObject(action.Key, action.Bucket, action.S3Class); err != nil {
			log.Fatal(err)
		}
	case GetAction:
		log.Printf("Attempting to restore %s", action.Key)
		exists, err := ObjectExists(action.Key, action.Bucket)
		if err != nil {
			log.Fatal(err)
		}
		// Get and and unzip
		var filename string
		if exists {
			log.Print("Cache hit, starting download")
			filename = action.Key
		} else {
			log.Printf("No caches found for the following key: %s", action.Key)
			log.Printf("Querying for cache matching default key: %s", action.DefaultKey)
			filename, err = GetLatestObject(action.DefaultKey, action.Bucket)
			if err != nil {
				log.Print(err)
				log.Print("Skipping cache download")
				return
			}
			log.Printf("Defaulting to latest similar key: %s", filename)
		}
		err = GetObject(filename, action.Bucket)
		if err != nil {
			log.Fatal(err)
		}

		if err := Unzip(filename); err != nil {
			log.Printf("Failed to unzip %s", filename)
			log.Fatal(err)
		}
	case DeleteAction:
		if err := DeleteObject(action.Key, action.Bucket); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("Action \"%s\" is not allowed. Valid options are: [%s, %s, %s]", act, PutAction, DeleteAction, GetAction)
	}
}
