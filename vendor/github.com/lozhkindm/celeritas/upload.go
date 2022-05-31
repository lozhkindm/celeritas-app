package celeritas

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"

	"github.com/lozhkindm/celeritas/filesystem"

	"github.com/gabriel-vasile/mimetype"
)

func (c *Celeritas) UploadFile(r *http.Request, field, dst string, fs filesystem.FileSystem) error {
	filename, err := c.getFileToUpload(r, field)
	defer func() {
		if err := os.Remove(filename); err != nil {
			c.ErrorLog.Println(err)
		}
	}()
	if err != nil {
		return err
	}
	if fs != nil {
		if err := fs.Put(filename, dst); err != nil {
			return err
		}
	} else {
		if err := os.Rename(filename, path.Join(dst, path.Base(filename))); err != nil {
			return err
		}
	}
	return nil
}

func (c *Celeritas) getFileToUpload(r *http.Request, field string) (string, error) {
	if err := r.ParseMultipartForm(c.config.upload.maxSize); err != nil {
		return "", err
	}

	file, header, err := r.FormFile(field)
	if err != nil {
		return "", err
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

	mime, err := mimetype.DetectReader(file)
	if err != nil {
		return "", err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return "", err
	}
	if !c.isValidMime(mime.String()) {
		return "", errors.New("invalid file type")
	}

	dst, err := os.Create(fmt.Sprintf("./tmp/%s", header.Filename))
	if err != nil {
		return "", err
	}
	defer func(dst *os.File) {
		_ = dst.Close()
	}(dst)

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("./tmp/%s", header.Filename), nil
}

func (c *Celeritas) isValidMime(mime string) bool {
	for _, v := range c.config.upload.allowedMimes {
		if mime == v {
			return true
		}
	}
	return false
}
