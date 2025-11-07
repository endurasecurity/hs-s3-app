package handlers

import (
	"html/template"
	"net/http"

	"hs-s3-app/storage"
)

// HomeHandler displays the dashboard
func HomeHandler(store *storage.MemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get statistics
		stats := store.GetStats()

		// Get recent AARs (all of them)
		recentAARs := store.GetAll()

		// Prepare template data
		data := map[string]interface{}{
			"Title":      "Dashboard",
			"Stats":      stats,
			"RecentAARs": recentAARs,
		}

		// Parse and execute templates
		tmpl, err := template.ParseFiles(
			"templates/layout.html",
			"templates/dashboard.html",
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
