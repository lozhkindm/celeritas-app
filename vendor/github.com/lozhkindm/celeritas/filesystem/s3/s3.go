package s3

import (
	"bytes"
	"net/http"
	"os"
	"path"

	"github.com/lozhkindm/celeritas/filesystem"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3 struct {
	Key      string
	Secret   string
	Region   string
	Endpoint string
	Bucket   string
}

func (s *S3) Put(filename, folder string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	fileinfo, err := file.Stat()
	if err != nil {
		return err
	}
	bts := make([]byte, fileinfo.Size())
	if _, err := file.Read(bts); err != nil {
		return err
	}
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(s.Endpoint),
		Region:      aws.String(s.Region),
		Credentials: credentials.NewStaticCredentials(s.Key, s.Secret, ""),
	}))
	_, err = s3manager.NewUploader(sess).Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(path.Join(folder, path.Base(filename))),
		Body:        bytes.NewReader(bts),
		ACL:         aws.String("public-read"),
		ContentType: aws.String(http.DetectContentType(bts)),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *S3) Get(dst string, items ...string) error {
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(s.Endpoint),
		Region:      aws.String(s.Region),
		Credentials: credentials.NewStaticCredentials(s.Key, s.Secret, ""),
	}))
	for _, item := range items {
		err := func() error {
			file, err := os.Create(path.Join(dst, path.Base(item)))
			if err != nil {
				return err
			}
			defer func() {
				_ = file.Close()
			}()
			input := &s3.GetObjectInput{
				Bucket: aws.String(s.Bucket),
				Key:    aws.String(item),
			}
			if _, err := s3manager.NewDownloader(sess).Download(file, input); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *S3) List(prefix string) ([]filesystem.ListEntry, error) {
	var entries []filesystem.ListEntry
	if prefix == "/" {
		prefix = ""
	}
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(s.Endpoint),
		Region:      aws.String(s.Region),
		Credentials: credentials.NewStaticCredentials(s.Key, s.Secret, ""),
	}))
	result, err := s3.New(sess).ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}
	for _, content := range result.Contents {
		entries = append(entries, filesystem.ListEntry{
			Etag:         *content.ETag,
			LastModified: *content.LastModified,
			Key:          *content.Key,
			Size:         float64(*content.Size) / 1024 / 1024,
		})
	}
	return entries, nil
}

func (s *S3) Delete(toDelete []string) (bool, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(s.Endpoint),
		Region:      aws.String(s.Region),
		Credentials: credentials.NewStaticCredentials(s.Key, s.Secret, ""),
	}))
	var objects []*s3.ObjectIdentifier
	for _, file := range toDelete {
		objects = append(objects, &s3.ObjectIdentifier{Key: aws.String(file)})
	}
	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(s.Bucket),
		Delete: &s3.Delete{
			Objects: objects,
			Quiet:   aws.Bool(false),
		},
	}
	if _, err := s3.New(sess).DeleteObjects(input); err != nil {
		return false, err
	}
	return true, nil
}
