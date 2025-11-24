package webapp

import (
	"net/http"

	"github.com/ya-breeze/diary.be/pkg/utils"
)

func (r *WebAppRouter) aboutHandler(w http.ResponseWriter, req *http.Request) {
	tmpl, err := r.loadTemplates()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := utils.CreateTemplateData(req, "about")

	if err := tmpl.ExecuteTemplate(w, "about.tpl", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
