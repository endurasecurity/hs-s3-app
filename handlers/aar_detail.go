package handlers

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"time"

	"hs-s3-app/s3"
	"hs-s3-app/storage"
)

// ViewAARHandler displays the detail view of an AAR
func ViewAARHandler(store *storage.MemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		aarID := r.URL.Query().Get("id")
		if aarID == "" {
			http.Error(w, "AAR ID is required", http.StatusBadRequest)
			return
		}

		aar, err := store.GetByID(aarID)
		if err != nil {
			http.Error(w, "AAR not found", http.StatusNotFound)
			return
		}

		// Prepare template data
		data := map[string]interface{}{
			"Title": aar.OperationName,
			"AAR":   aar,
		}

		// Parse and execute templates
		funcMap := template.FuncMap{
			"divf": func(a int64, b float64) float64 {
				return float64(a) / b
			},
		}

		tmpl, err := template.New("layout.html").Funcs(funcMap).ParseFiles(
			"templates/layout.html",
			"templates/detail.html",
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

// DownloadAttachmentHandler handles downloading attachments from S3
func DownloadAttachmentHandler(store *storage.MemoryStore, s3Client *s3.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		aarID := r.URL.Query().Get("id")
		filename := r.URL.Query().Get("file")

		if aarID == "" || filename == "" {
			http.Error(w, "AAR ID and filename are required", http.StatusBadRequest)
			return
		}

		// Get AAR to verify it exists and find the attachment
		aar, err := store.GetByID(aarID)
		if err != nil {
			http.Error(w, "AAR not found", http.StatusNotFound)
			return
		}

		// Find the attachment
		var s3Key string
		var contentType string
		for _, att := range aar.Attachments {
			if att.Filename == filename {
				s3Key = att.S3Key
				contentType = att.ContentType
				break
			}
		}

		if s3Key == "" {
			http.Error(w, "Attachment not found", http.StatusNotFound)
			return
		}

		// If S3 client is not configured, return error
		if s3Client == nil {
			http.Error(w, "S3 client not configured", http.StatusInternalServerError)
			return
		}

		// Download from S3
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		reader, err := s3Client.DownloadFile(ctx, s3Key)
		if err != nil {
			http.Error(w, "Error downloading file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer reader.Close()

		// Set headers for download
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")

		// Stream the file to the response
		if _, err := io.Copy(w, reader); err != nil {
			http.Error(w, "Error streaming file: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
