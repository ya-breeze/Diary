package webapp

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/ya-breeze/diary.be/pkg/generated/goserver"
	"github.com/ya-breeze/diary.be/pkg/server/common"
	"github.com/ya-breeze/diary.be/pkg/utils"
)

func (r *WebAppRouter) homeHandler(w http.ResponseWriter, req *http.Request) {
	// Load Go templates with custom functions and template inheritance
	tmpl, err := r.loadTemplates()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize template data with common request information
	data := utils.CreateTemplateData(req, "home")

	// Authenticate user and extract family ID from cookie
	familyID, err := r.ValidateFamilyID(tmpl, w, req)
	if err != nil {
		r.logger.Error("Failed to get family ID from cookie", "error", err)
		return
	}
	data["FamilyID"] = familyID.String()

	// Determine target date: use query parameter or default to current date
	date := req.URL.Query().Get("date")
	if date == "" {
		date = utils.GetCurrentDate()
	}

	// Fetch diary entry data and populate template with content
	if err := r.populateItemsData(data, familyID, date, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add layout toggle configuration and feature flags to template data
	r.addLayoutTemplateData(data, req)

	// Render the home template with all collected data
	r.renderTemplate(w, tmpl, data)
}

// populateItemsData fetches items data and populates the template data
func (r *WebAppRouter) populateItemsData(data map[string]any, familyID uuid.UUID, date string, req *http.Request) error {
	// Create context with family ID for the items service
	ctx := context.WithValue(req.Context(), common.FamilyIDKey, familyID)

	// Use the items service to get items (new API signature with search parameters)
	// For home page, we use date filter for backward compatibility
	response, err := r.itemsService.GetItems(ctx, date, "", "")
	if err != nil {
		r.logger.Error("Failed to get items from service", "error", err, "date", date)
		return err
	}

	if response.Code != 200 {
		r.logger.Error("Items service returned non-200 status", "code", response.Code, "date", date)
		return errors.New("failed to get items")
	}

	itemsListResponse, ok := response.Body.(goserver.ItemsListResponse)
	if !ok {
		r.logger.Error("Failed to cast response body to ItemsListResponse")
		return errors.New("internal server error")
	}

	// Handle backward compatibility: for home page with date filter, we expect 0 or 1 item
	var itemsResponse goserver.ItemsResponse
	if len(itemsListResponse.Items) > 0 {
		// Use the first (and should be only) item for the specific date
		itemsResponse = itemsListResponse.Items[0]
	} else {
		// Create empty item for the requested date (backward compatibility)
		parsedDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
		emptyBody := ""
		emptyTags := []string{}
		itemsResponse = goserver.ItemsResponse{
			Date:  openapi_types.Date{Time: parsedDate},
			Title: "",
			Body:  &emptyBody,
			Tags:  &emptyTags,
		}
		// For empty items, we need to manually add navigation dates
		// since the service doesn't populate them for non-existent items
		if previousDate, err := r.db.GetPreviousDate(familyID, date); err == nil {
			t, _ := time.Parse("2006-01-02", previousDate)
			d := openapi_types.Date{Time: t}
			itemsResponse.PreviousDate = &d
		}
		if nextDate, err := r.db.GetNextDate(familyID, date); err == nil {
			t, _ := time.Parse("2006-01-02", nextDate)
			d := openapi_types.Date{Time: t}
			itemsResponse.NextDate = &d
		}
	}

	bodyStr := ""
	if itemsResponse.Body != nil {
		bodyStr = *itemsResponse.Body
	}
	var tags []string
	if itemsResponse.Tags != nil {
		tags = *itemsResponse.Tags
	}

	// Convert the service response to template data (maintaining existing structure)
	data["item"] = map[string]any{
		"Date":  itemsResponse.Date.Time.Format("2006-01-02"),
		"Title": itemsResponse.Title,
		"Body":  bodyStr,
		"Tags":  tags,
	}

	body := markdown.ToHTML([]byte(bodyStr), nil, utils.NewImagePrefixRenderer("/web/assets/"))
	//nolint:gosec // this is safe
	data["body"] = template.HTML(string(body))

	// Add navigation dates from the service response
	if itemsResponse.PreviousDate != nil {
		data["previousDate"] = itemsResponse.PreviousDate.Time.Format("2006-01-02")
	}
	if itemsResponse.NextDate != nil {
		data["nextDate"] = itemsResponse.NextDate.Time.Format("2006-01-02")
	}

	return nil
}

func (r *WebAppRouter) addLayoutTemplateData(data map[string]any, req *http.Request) {
	// Feature flags for conditional template rendering
	// These allow templates to show/hide layout controls based on capabilities
	data["LayoutToggleEnabled"] = true
	data["JavaScriptEnabled"] = true // Initial assumption, client-side JS will update this

	// Default layout preference for server-side rendering
	// This ensures consistent initial state before client-side preferences load
	data["DefaultLayout"] = "narrow"

	// Layout configuration object for JavaScript initialization
	// Provides centralized configuration that can be accessed by client-side code
	data["LayoutConfig"] = map[string]any{
		"FullWidthPercent":   100, // Full layout mode image width percentage
		"NarrowWidthPercent": 30,  // Narrow layout mode image width percentage
		"TransitionDuration": 300, // CSS transition duration in milliseconds
	}

	// User agent analysis for responsive behavior hints
	userAgent := req.Header.Get("User-Agent")
	data["UserAgent"] = userAgent
	data["IsMobile"] = utils.IsMobile(userAgent) // Server-side mobile detection

	// Cache busting timestamp for static assets
	// Helps ensure users get updated CSS/JS files after deployments
	data["Timestamp"] = time.Now().Unix()
}

// renderTemplate renders the template with the provided data
func (r *WebAppRouter) renderTemplate(w http.ResponseWriter, tmpl *template.Template, data map[string]any) {
	data["Template"] = "home.tpl"

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
