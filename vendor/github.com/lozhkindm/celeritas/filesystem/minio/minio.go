package minio

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/lozhkindm/celeritas/filesystem"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	Endpoint string
	Key      string
	Secret   string
	UseSSL   bool
	Region   string
	Bucket   string
}

func (m *Minio) getCredentials() (*minio.Client, error) {
	return minio.New(m.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(m.Key, m.Secret, ""),
		Secure: m.UseSSL,
	})
}

func (m *Minio) Put(filename, folder string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectName := path.Base(filename)
	client, err := m.getCredentials()
	if err != nil {
		return err
	}
	_, err = client.FPutObject(ctx, m.Bucket, fmt.Sprintf("%s/%s", folder, objectName), filename, minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (m *Minio) Get(dst string, items ...string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := m.getCredentials()
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := client.FGetObject(ctx, m.Bucket, item, fmt.Sprintf("%s/%s", dst, path.Base(item)), minio.GetObjectOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func (m *Minio) List(prefix string) ([]filesystem.ListEntry, error) {
	var entries []filesystem.ListEntry
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := m.getCredentials()
	if err != nil {
		return nil, err
	}
	objectCh := client.ListObjects(ctx, m.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})
	for objectInfo := range objectCh {
		if objectInfo.Err != nil {
			return nil, objectInfo.Err
		}
		if !strings.HasPrefix(objectInfo.Key, ".") {
			entries = append(entries, filesystem.ListEntry{
				Etag:         objectInfo.ETag,
				LastModified: objectInfo.LastModified,
				Key:          objectInfo.Key,
				Size:         float64(objectInfo.Size) / 1024 / 1024, // MB
			})
		}
	}
	return entries, nil
}

func (m *Minio) Delete(toDelete []string) (bool, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := m.getCredentials()
	if err != nil {
		return false, err
	}
	for _, item := range toDelete {
		err := client.RemoveObject(ctx, m.Bucket, item, minio.RemoveObjectOptions{
			GovernanceBypass: true,
		})
		if err != nil {
			return false, err
		}
	}
	return true, nil
}
