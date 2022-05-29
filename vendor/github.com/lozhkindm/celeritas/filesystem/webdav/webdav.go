package webdav

import (
	"io"
	"os"
	"path"
	"strings"

	"github.com/lozhkindm/celeritas/filesystem"

	"github.com/studio-b12/gowebdav"
)

type WebDAV struct {
	Host     string
	User     string
	Password string
}

func (w *WebDAV) Put(filename, folder string) error {
	client := gowebdav.NewClient(w.Host, w.User, w.Password)
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	var bytes []byte
	if _, err := file.Read(bytes); err != nil {
		return err
	}
	if err := client.Write(path.Join(folder, path.Base(filename)), bytes, 0644); err != nil {
		return err
	}
	//if err := client.WriteStream(path.Join(folder, path.Base(filename)), file, 0644); err != nil {
	//	return err
	//}
	return nil
}

func (w *WebDAV) Get(dst string, items ...string) error {
	client := gowebdav.NewClient(w.Host, w.User, w.Password)
	for _, item := range items {
		err := func() error {
			dstFile, err := os.Create(path.Join(dst, path.Base(item)))
			if err != nil {
				return err
			}
			defer func() {
				_ = dstFile.Close()
			}()
			reader, err := client.ReadStream(item)
			if err != nil {
				return err
			}
			if _, err := io.Copy(dstFile, reader); err != nil {
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

func (w *WebDAV) List(prefix string) ([]filesystem.ListEntry, error) {
	var entries []filesystem.ListEntry
	client := gowebdav.NewClient(w.Host, w.User, w.Password)
	files, err := client.ReadDir(prefix)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") {
			entries = append(entries, filesystem.ListEntry{
				LastModified: file.ModTime(),
				Key:          file.Name(),
				Size:         float64(file.Size()) / 1024 / 1024, // MB
				IsDir:        file.IsDir(),
			})
		}
	}
	return entries, nil
}

func (w *WebDAV) Delete(toDelete []string) (bool, error) {
	client := gowebdav.NewClient(w.Host, w.User, w.Password)
	for _, file := range toDelete {
		if err := client.Remove(file); err != nil {
			return false, err
		}
	}
	return true, nil
}
