package celeritas

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
)

func (c *Celeritas) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1048576 // 1 MB

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(data); err != nil {
		return err
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body must only have a single json value")
	}
	return nil
}

func (c *Celeritas) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	res, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(res); err != nil {
		return err
	}
	return nil
}

func (c *Celeritas) WriteXML(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	res, err := xml.MarshalIndent(data, "", "   ")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)

	if _, err := w.Write(res); err != nil {
		return err
	}
	return nil
}

func (c *Celeritas) DownloadFile(w http.ResponseWriter, r *http.Request, pathToFile, filename string) {
	fullPath := path.Join(pathToFile, filename)
	file := filepath.Clean(fullPath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; file=\"%s\"", filename))
	http.ServeFile(w, r, file)
}

func (c *Celeritas) NotFound(w http.ResponseWriter) {
	c.ErrorStatus(w, http.StatusNotFound)
}

func (c *Celeritas) InternalError(w http.ResponseWriter) {
	c.ErrorStatus(w, http.StatusInternalServerError)
}

func (c *Celeritas) Unauthorized(w http.ResponseWriter) {
	c.ErrorStatus(w, http.StatusUnauthorized)
}

func (c *Celeritas) Forbidden(w http.ResponseWriter) {
	c.ErrorStatus(w, http.StatusForbidden)
}

func (c *Celeritas) ErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
