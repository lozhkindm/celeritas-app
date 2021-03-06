package handlers

import (
	"fmt"
	"net/http"

	"myapp/data"

	"github.com/CloudyKit/jet/v6"
)

func (h *Handlers) ShowForm(w http.ResponseWriter, r *http.Request) {
	vars := make(jet.VarMap)
	validator := h.App.Validator(nil)
	vars.Set("validator", validator)
	vars.Set("user", data.User{})

	if err := h.render(w, r, "form", vars, nil); err != nil {
		h.App.ErrorLog.Println(err)
	}
}

func (h *Handlers) SubmitForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.App.ErrorLog.Println(err)
		return
	}

	validator := h.App.Validator(nil)
	validator.Required(r, "first_name", "last_name", "email")
	validator.IsEmail("email", r.Form.Get("email"))
	validator.Check(len(r.Form.Get("first_name")) > 1, "first_name", "First name must contain at least 1 character")
	validator.Check(len(r.Form.Get("last_name")) > 1, "last_name", "Last name must contain at least 1 character")

	if !validator.IsValid() {
		vars := make(jet.VarMap)
		vars.Set("validator", validator)

		var user data.User
		user.FirstName = r.Form.Get("first_name")
		user.LastName = r.Form.Get("last_name")
		user.Email = r.Form.Get("email")
		vars.Set("user", user)

		if err := h.render(w, r, "form", vars, nil); err != nil {
			h.App.ErrorLog.Println(err)
		}

		return
	}

	fmt.Fprint(w, "valid data")
}
