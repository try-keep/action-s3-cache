package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Zip - Create .zip file and add dirs and files that match glob patterns
func Zip(filename string, artifacts []string) error {
	start := time.Now()
	log.Printf("Starting to zip: %s", filename)
	// tar + gzip
	var buf bytes.Buffer
	zr := gzip.NewWriter(&buf)
	tw := tar.NewWriter(zr)

	for _, pattern := range artifacts {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		for _, match := range matches {
			// walk through every file in the folder
			filepath.Walk(match, func(file string, fi os.FileInfo, err error) error {
				// generate tar header
				header, err := tar.FileInfoHeader(fi, file)
				if err != nil {
					return err
				}

				// must provide real name
				// (see https://golang.org/src/archive/tar/common.go?#L626)
				header.Name = filepath.ToSlash(file)

				// write header
				if err := tw.WriteHeader(header); err != nil {
					return err
				}
				// if not a dir, write file content
				if !fi.IsDir() {
					data, err := os.Open(file)
					if err != nil {
						return err
					}
					if _, err := io.Copy(tw, data); err != nil {
						return err
					}
					return nil
				}
				return nil
			})
		}
	}

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}
	// produce gzip
	if err := zr.Close(); err != nil {
		return err
	}

	// write the .tar.gzip
	fileToWrite, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, os.FileMode(0600))
	if err != nil {
		log.Printf("Failed to open file \"%s\"", filename)
		panic(err)
	}
	if _, err := io.Copy(fileToWrite, &buf); err != nil {
		log.Printf("Failed copying buffer to open file %s", filename)
		panic(err)
	}
	elapsed := time.Since(start)
	file, err := fileToWrite.Stat()
	if err != nil {
		panic(err)
	}

	log.Printf("Successfully zipped %v in %s!", getReadableBytes(file.Size()), elapsed)
	return os.Chmod(filename, 0777)
}

// Unzip - Unzip all files and directories inside .zip file
func Unzip(filename string) error {
	start := time.Now()
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
		target := filepath.ToSlash(header.Name)

		if !hasValidRelPath(header.Name) {
			return fmt.Errorf("TAR contained invalid name, %q", target)
		}

		if header.Typeflag == tar.TypeReg {
			// Create the directory that contains it
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Printf("Failed to create directory %s", dir)
			}

			// Write the file
			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				log.Printf("Failed creating %s", target)
				panic(err)
			}
			// Copy over contents
			if _, err := io.Copy(fileToWrite, tarReader); err != nil {
				log.Printf("Failed copying contents to %s", target)
				panic(err)
			}
			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			fileToWrite.Close()
		}
	}
	elapsed := time.Since(start)
	log.Printf("Successfully unzipped: %s in %s", filename, elapsed)
	return nil
}

func getReadableBytes(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func hasValidRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}
