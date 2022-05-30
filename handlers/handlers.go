package handlers

import (
	"fmt"
	"github.com/lozhkindm/celeritas/filesystem/s3"
	"github.com/lozhkindm/celeritas/filesystem/webdav"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	"myapp/data"

	"github.com/CloudyKit/jet/v6"
	"github.com/lozhkindm/celeritas"
	"github.com/lozhkindm/celeritas/filesystem"
	"github.com/lozhkindm/celeritas/filesystem/minio"
	"github.com/lozhkindm/celeritas/filesystem/sftp"
)

type Handlers struct {
	App    *celeritas.Celeritas
	Models data.Models
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	defer h.App.LoadTime(time.Now())
	if err := h.render(w, r, "home", nil, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) GoPage(w http.ResponseWriter, r *http.Request) {
	h.App.Render.Renderer = "go"
	if err := h.render(w, r, "home", nil, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) JetPage(w http.ResponseWriter, r *http.Request) {
	h.App.Render.Renderer = "jet"
	if err := h.render(w, r, "jet-template", nil, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) Sessions(w http.ResponseWriter, r *http.Request) {
	myData := "bar"
	h.sessionPut(r.Context(), "foo", myData)
	myValue := h.sessionGetString(r.Context(), "foo")
	vars := make(jet.VarMap)
	vars.Set("foo", myValue)

	if err := h.render(w, r, "sessions", vars, nil); err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) JSON(w http.ResponseWriter, _ *http.Request) {
	var payload struct {
		ID      int64    `json:"id"`
		Name    string   `json:"name"`
		Hobbies []string `json:"hobbies"`
	}

	payload.ID = 123
	payload.Name = "Ignat Senkin"
	payload.Hobbies = []string{"CS", "Formula", "Cards"}

	if err := h.App.WriteJSON(w, http.StatusOK, payload); err != nil {
		h.App.ErrorLog.Println(err)
	}
}

func (h *Handlers) XML(w http.ResponseWriter, _ *http.Request) {
	type payload struct {
		ID      int64    `xml:"id"`
		Name    string   `xml:"name"`
		Hobbies []string `xml:"hobbies>hobby"`
	}
	var pl payload

	pl.ID = 123
	pl.Name = "Ignat Senkin"
	pl.Hobbies = []string{"CS", "Formula", "Cards"}

	if err := h.App.WriteXML(w, http.StatusOK, pl); err != nil {
		h.App.ErrorLog.Println(err)
	}
}

func (h *Handlers) DownloadFile(w http.ResponseWriter, r *http.Request) {
	h.App.DownloadFile(w, r, "./public/images", "celeritas.jpg")
}

func (h *Handlers) TestCrypto(w http.ResponseWriter, _ *http.Request) {
	plaintext := "Hello, world"
	_, _ = fmt.Fprintf(w, "plaintext: %s\n", plaintext)

	encrypted, err := h.encrypt(plaintext)
	if err != nil {
		h.App.ErrorLog.Println(err)
		h.App.InternalError(w)
		return
	}

	_, _ = fmt.Fprintf(w, "encrypted: %s\n", encrypted)

	decrypted, err := h.decrypt(encrypted)
	if err != nil {
		h.App.ErrorLog.Println(err)
		h.App.InternalError(w)
		return
	}

	_, _ = fmt.Fprintf(w, "decrypted: %s", decrypted)
}

func (h *Handlers) ListFileSystems(w http.ResponseWriter, r *http.Request) {
	var (
		fsType  string
		curPath = "/"
		err     error
		fs      filesystem.FileSystem
		entries []filesystem.ListEntry
	)
	if ft := r.URL.Query().Get("fs-type"); ft != "" {
		fsType = ft
	}
	if cp := r.URL.Query().Get("cur-path"); cp != "" {
		if cp, err = url.QueryUnescape(cp); err != nil {
			h.App.ErrorLog.Println(err)
			return
		}
		curPath = cp
	}
	if fsType != "" {
		switch fsType {
		case "MINIO":
			f := h.App.FileSystems["MINIO"].(minio.Minio)
			fs = &f
		case "SFTP":
			f := h.App.FileSystems["SFTP"].(sftp.SFTP)
			fs = &f
		case "WEBDAV":
			f := h.App.FileSystems["WEBDAV"].(webdav.WebDAV)
			fs = &f
		case "S3":
			f := h.App.FileSystems["S3"].(s3.S3)
			fs = &f
		}
		if entries, err = fs.List(curPath); err != nil {
			h.App.ErrorLog.Println(err)
			return
		}
	}
	vars := make(jet.VarMap)
	vars.Set("list", entries)
	vars.Set("fs_type", fsType)
	vars.Set("curPath", curPath)
	if err = h.render(w, r, "list-filesystems", vars, nil); err != nil {
		h.App.ErrorLog.Println(err)
		return
	}
}

func (h *Handlers) FormUploadFileToFileSystem(w http.ResponseWriter, r *http.Request) {
	fsType := r.URL.Query().Get("type")
	vars := make(jet.VarMap)
	vars.Set("fs_type", fsType)
	if err := h.render(w, r, "upload", vars, nil); err != nil {
		h.App.ErrorLog.Println(err)
		return
	}
}

func (h *Handlers) PostUploadFileToFileSystem(w http.ResponseWriter, r *http.Request) {
	filename, err := getFileToUpload(r, "formFile")
	if err != nil {
		h.App.ErrorLog.Println(err)
		h.App.InternalError(w)
		return
	}

	var fs filesystem.FileSystem
	tp := r.Form.Get("upload-type")

	switch tp {
	case "MINIO":
		f := h.App.FileSystems["MINIO"].(minio.Minio)
		fs = &f
	case "SFTP":
		f := h.App.FileSystems["SFTP"].(sftp.SFTP)
		fs = &f
	case "WEBDAV":
		f := h.App.FileSystems["WEBDAV"].(webdav.WebDAV)
		fs = &f
	case "S3":
		f := h.App.FileSystems["S3"].(s3.S3)
		fs = &f
	}

	if err := fs.Put(filename, ""); err != nil {
		h.App.ErrorLog.Println(err)
		h.App.InternalError(w)
		return
	}
	h.App.Session.Put(r.Context(), "flash", "File uploaded")
	http.Redirect(w, r, fmt.Sprintf("/files/upload?type=%s", tp), http.StatusSeeOther)
}

func (h *Handlers) DeleteFromFileSystem(w http.ResponseWriter, r *http.Request) {
	var (
		fs       filesystem.FileSystem
		fsType   string
		filename string
	)
	fsType = r.URL.Query().Get("fs_type")
	filename = r.URL.Query().Get("file")

	switch fsType {
	case "MINIO":
		f := h.App.FileSystems["MINIO"].(minio.Minio)
		fs = &f
	case "SFTP":
		f := h.App.FileSystems["SFTP"].(sftp.SFTP)
		fs = &f
	case "WEBDAV":
		f := h.App.FileSystems["WEBDAV"].(webdav.WebDAV)
		fs = &f
	case "S3":
		f := h.App.FileSystems["S3"].(s3.S3)
		fs = &f
	}

	deleted, err := fs.Delete([]string{filename})
	if err != nil {
		h.App.ErrorLog.Println(err)
		h.App.InternalError(w)
		return
	}
	if deleted {
		h.App.Session.Put(r.Context(), "flash", fmt.Sprintf("%s was deleted", filename))
	}

	http.Redirect(w, r, fmt.Sprintf("/list-fs?fs-type=%s", fsType), http.StatusSeeOther)
}

func (h *Handlers) FormGenericUpload(w http.ResponseWriter, r *http.Request) {
	if err := h.render(w, r, "generic-upload", nil, nil); err != nil {
		h.App.ErrorLog.Println(err)
		return
	}
}

func (h *Handlers) PostGenericUpload(w http.ResponseWriter, r *http.Request) {
	if err := h.App.UploadFile(r, "formFile", "", &h.App.SFTP); err != nil {
		h.App.ErrorLog.Println(err)
		h.App.Session.Put(r.Context(), "error", err.Error())
	} else {
		h.App.Session.Put(r.Context(), "flash", "File uploaded")
	}
	http.Redirect(w, r, "/upload", http.StatusSeeOther)
}

func getFileToUpload(r *http.Request, key string) (string, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return "", err
	}

	file, header, err := r.FormFile(key)
	if err != nil {
		return "", err
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

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
