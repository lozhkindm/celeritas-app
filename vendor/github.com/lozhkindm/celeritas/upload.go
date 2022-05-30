package celeritas

import (
	"errors"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"

	"github.com/lozhkindm/celeritas/filesystem"
)

func (c *Celeritas) UploadFile(r *http.Request, field, dst string, fs filesystem.FileSystem) error {
	filename, err := getFileToUpload(r, field)
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

func getFileToUpload(r *http.Request, field string) (string, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
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
	if !isValidMime(mime.String()) {
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

func isValidMime(mime string) bool {
	valid := []string{"image/gif", "image/jpeg", "image/png", "application/pdf"}
	for _, v := range valid {
		if mime == v {
			return true
		}
	}
	return false
}
