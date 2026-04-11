package webapp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/ya-breeze/diary.be/pkg/database"
	"github.com/ya-breeze/diary.be/pkg/database/models"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/common"
	"github.com/ya-breeze/diary.be/pkg/utils"
)

func (r *WebAppRouter) editHandler(w http.ResponseWriter, req *http.Request) {
	tmpl, err := r.loadTemplates()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := utils.CreateTemplateData(req, "edit")

	familyID, err := r.ValidateFamilyID(tmpl, w, req)
	if err != nil {
		r.logger.Error("Failed to get family ID from cookie", "error", err)
		return
	}
	data["FamilyID"] = familyID.String()

	date := req.URL.Query().Get("date")
	if date == "" {
		date = utils.GetCurrentDate()
	}
	item, err := r.db.GetItem(familyID, date)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			r.logger.Error("Failed to get item", "error", err, "date", date, "familyID", familyID)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		item = &models.Item{
			Date:  date,
			Title: "",
			Body:  "",
		}
	}
	data["item"] = item
	data["assets"] = utils.GetAssetsFromMarkdown(item.Body)

	templateName := "edit.tpl"
	if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
		r.logger.Warn("failed to execute template", "error", err, "template", templateName)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (r *WebAppRouter) saveHandler(w http.ResponseWriter, req *http.Request) {
	tmpl, err := r.loadTemplates()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := utils.CreateTemplateData(req, "edit")

	familyID, err := r.ValidateFamilyID(tmpl, w, req)
	if err != nil {
		r.logger.Error("Failed to get family ID from cookie", "error", err)
		return
	}
	data["FamilyID"] = familyID.String()

	date := req.FormValue("date")
	if date == "" {
		http.Error(w, "Date is required", http.StatusBadRequest)
		return
	}

	// Build the API request and call the Items API service instead of writing to DB directly
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		http.Error(w, "invalid date format", http.StatusBadRequest)
		return
	}
	body := req.FormValue("body")
	tags := strings.Split(req.FormValue("tags"), ",")
	itemsRequest := goserver.ItemsRequest{
		Date:  openapi_types.Date{Time: parsedDate},
		Title: req.FormValue("title"),
		Body:  &body,
		Tags:  &tags,
	}

	// Ensure the service can read the family ID from context (the API service expects it there)
	ctx := context.WithValue(req.Context(), common.FamilyIDKey, familyID)

	implResp, svcErr := r.itemsService.PutItems(ctx, itemsRequest)
	if svcErr != nil {
		r.logger.Error("Items service returned error", "error", svcErr)
		http.Error(w, svcErr.Error(), http.StatusInternalServerError)
		return
	}

	// Handle non-OK response codes from the service
	if implResp.Code >= 400 {
		r.logger.Error("Items service returned non-OK code", "code", implResp.Code)
		http.Error(w, http.StatusText(implResp.Code), implResp.Code)
		return
	}

	// On success redirect to the saved date
	http.Redirect(w, req, "/?date="+date, http.StatusSeeOther)
}
