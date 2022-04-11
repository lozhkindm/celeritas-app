package celeritas

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

type Validation struct {
	Data   url.Values
	Errors map[string]string
}

func (c *Celeritas) Validator(data url.Values) *Validation {
	return &Validation{
		Data:   data,
		Errors: make(map[string]string),
	}
}

func (v *Validation) IsValid() bool {
	return len(v.Errors) == 0
}

func (v *Validation) AddError(field, message string) {
	if _, ok := v.Errors[field]; !ok {
		v.Errors[field] = message
	}
}

func (v *Validation) Has(field string, r *http.Request) bool {
	val := r.Form.Get(field)
	return val != ""
}

func (v *Validation) Required(r *http.Request, fields ...string) {
	for _, f := range fields {
		val := r.Form.Get(f)
		if strings.TrimSpace(val) == "" {
			v.AddError(f, fmt.Sprintf("Field '%s' is required", f))
		}
	}
}

func (v *Validation) Check(cond bool, field, message string) {
	if !cond {
		v.AddError(field, message)
	}
}

func (v *Validation) IsEmail(field, value string) {
	if !govalidator.IsEmail(value) {
		v.AddError(field, fmt.Sprintf("Field '%s' must be a valid email address", field))
	}
}

func (v *Validation) IsInt(field, value string) {
	if _, err := strconv.Atoi(value); err != nil {
		v.AddError(field, fmt.Sprintf("Field '%s' must be an integer"))
	}
}

func (v *Validation) IsFloat(field, value string) {
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		v.AddError(field, fmt.Sprintf("Field '%s' must be a float"))
	}
}

func (v *Validation) IsDateISO(field, value string) {
	if _, err := time.Parse("2006-01-02", value); err != nil {
		v.AddError(field, fmt.Sprintf("Field '%s' must be a date in the form of YYYY-MM-DD"))
	}
}

func (v *Validation) NoSpaces(field, value string) {
	if govalidator.HasWhitespace(value) {
		v.AddError(field, fmt.Sprintf("Field '%s' must not contain whitespaces"))
	}
}
