package handlers

import (
	"html/template"
	"net/http"

	"hs-s3-app/models"
	"hs-s3-app/storage"
)

// ListAARHandler displays the list of AARs with search/filter capabilities
func ListAARHandler(store *storage.MemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get search parameters from query string
		operationName := r.URL.Query().Get("operation_name")
		unit := r.URL.Query().Get("unit")
		missionType := r.URL.Query().Get("mission_type")

		// Search AARs
		var aars []*models.AAR
		if operationName != "" || unit != "" || missionType != "" {
			aars = store.Search(operationName, unit, missionType)
		} else {
			aars = store.GetAll()
		}

		// Prepare template data
		data := map[string]interface{}{
			"Title": "Browse AARs",
			"AARs":  aars,
			"SearchParams": map[string]string{
				"OperationName": operationName,
				"Unit":          unit,
				"MissionType":   missionType,
			},
		}

		// Parse and execute templates
		tmpl, err := template.ParseFiles(
			"templates/layout.html",
			"templates/list.html",
		)
		if err != nil {
			http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
			http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
