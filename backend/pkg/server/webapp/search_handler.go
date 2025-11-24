package webapp

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/common"
	"github.com/ya-breeze/diary.be/pkg/utils"
)

func (r *WebAppRouter) searchHandler(w http.ResponseWriter, req *http.Request) {
	// Load Go templates with custom functions and template inheritance
	tmpl, err := r.loadTemplates()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize template data with common request information
	data := utils.CreateTemplateData(req, "search")

	// Authenticate user and extract user ID from session
	// This may redirect to login page if authentication fails
	userID, err := r.ValidateUserID(tmpl, w, req)
	if err != nil {
		r.logger.Error("Failed to get user ID from session", "error", err)
		return
	}
	data["UserID"] = userID

	// Extract search parameters from query string
	searchQuery := strings.TrimSpace(req.URL.Query().Get("search"))
	tagsParam := strings.TrimSpace(req.URL.Query().Get("tags"))
	dateParam := strings.TrimSpace(req.URL.Query().Get("date"))

	// Parse tags parameter (comma-separated)
	var searchTags []string
	if tagsParam != "" {
		for tag := range strings.SplitSeq(tagsParam, ",") {
			t := strings.TrimSpace(tag)
			if t != "" {
				searchTags = append(searchTags, t)
			}
		}
	}

	// Fetch search results and populate template with content
	if err := r.populateSearchData(data, userID, searchQuery, searchTags, dateParam, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add layout toggle configuration and feature flags to template data
	r.addLayoutTemplateData(data, req)

	// Set template name for search results
	data["Template"] = "search.tpl"

	// Render the search template with all collected data
	r.renderSearchTemplate(w, tmpl, data)
}

// populateSearchData fetches search results and populates the template data
//
//nolint:funlen // acceptable length for orchestrating search flow and template population
func (r *WebAppRouter) populateSearchData(
	data map[string]any,
	userID string,
	searchQuery string,
	searchTags []string,
	dateParam string,
	req *http.Request,
) error {
	// Create context with user ID for the items service
	ctx := context.WithValue(req.Context(), common.UserIDKey, userID)

	// Prepare search parameters
	tagsParam := strings.Join(searchTags, ",")

	// Use the items service to get search results
	response, err := r.itemsService.GetItems(ctx, dateParam, searchQuery, tagsParam)
	if err != nil {
		r.logger.Error(
			"Failed to get search results from service",
			"error", err,
			"searchQuery", searchQuery,
			"tags", tagsParam,
			"userID", userID,
		)
		return err
	}

	if response.Code != 200 {
		r.logger.Error(
			"Items service returned non-200 status",
			"code", response.Code,
			"searchQuery", searchQuery,
			"tags", tagsParam,
			"userID", userID,
		)
		return errors.New("failed to get search results")
	}

	itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
	if !ok {
		r.logger.Error("Failed to cast response body to ItemsListResponse")
		return errors.New("internal server error")
	}

	// Convert the service response to template data
	items := make([]map[string]any, len(itemsListResponse.Items))
	for i, item := range itemsListResponse.Items {
		items[i] = map[string]any{
			"Date":  item.Date,
			"Title": item.Title,
			"Body":  item.Body, // Keep original for truncation logic in template
			"Tags":  item.Tags,
		}
	}

	// Add search results to template data
	data["items"] = items
	data["totalCount"] = int(itemsListResponse.TotalCount)
	data["searchQuery"] = searchQuery
	data["searchTags"] = searchTags

	// Add search context information
	if searchQuery != "" {
		data["hasSearchQuery"] = true
	}
	if len(searchTags) > 0 {
		data["hasSearchTags"] = true
	}

	return nil
}

// renderSearchTemplate renders the search template with the provided data
func (r *WebAppRouter) renderSearchTemplate(w http.ResponseWriter, tmpl *template.Template, data map[string]any) {
	templateName, ok := data["Template"].(string)
	if !ok {
		r.logger.Error("Failed to assert template name")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
		r.logger.Warn("failed to execute template", "error", err, "template", templateName)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
