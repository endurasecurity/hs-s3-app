package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"hs-s3-app/models"
	"hs-s3-app/s3"
	"hs-s3-app/storage"
)

// CreateAARHandler handles displaying the creation form and processing submissions
func CreateAARHandler(store *storage.MemoryStore, s3Client *s3.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			showCreateForm(w, nil, "")
			return
		}

		if r.Method == http.MethodPost {
			handleCreateSubmission(w, r, store, s3Client)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func showCreateForm(w http.ResponseWriter, err error, aarID string) {
	data := map[string]interface{}{
		"Title": "Create AAR",
	}

	if err != nil {
		data["Error"] = err.Error()
	}

	if aarID != "" {
		data["Success"] = true
		data["AARID"] = aarID
	}

	tmpl, parseErr := template.ParseFiles(
		"templates/layout.html",
		"templates/create.html",
	)
	if parseErr != nil {
		http.Error(w, "Error loading template: "+parseErr.Error(), http.StatusInternalServerError)
		return
	}

	if execErr := tmpl.ExecuteTemplate(w, "layout.html", data); execErr != nil {
		http.Error(w, "Error rendering template: "+execErr.Error(), http.StatusInternalServerError)
	}
}

func handleCreateSubmission(w http.ResponseWriter, r *http.Request, store *storage.MemoryStore, s3Client *s3.Client) {
	// Parse multipart form with max 100MB
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		showCreateForm(w, fmt.Errorf("error parsing form: %w", err), "")
		return
	}

	// Generate AAR ID
	now := time.Now()
	aarID := fmt.Sprintf("AAR-%s-%04d", now.Format("20060102"), now.Unix()%10000)

	// Parse personnel count
	personnelCount := 0
	if pc := r.FormValue("personnel_count"); pc != "" {
		personnelCount, _ = strconv.Atoi(pc)
	}

	// Create AAR object
	aar := &models.AAR{
		ID:                   aarID,
		Classification:       r.FormValue("classification"),
		OperationName:        r.FormValue("operation_name"),
		DTG:                  r.FormValue("dtg"),
		UnitDesignation:      r.FormValue("unit_designation"),
		MissionType:          r.FormValue("mission_type"),
		Location:             r.FormValue("location"),
		DurationStart:        r.FormValue("duration_start"),
		DurationEnd:          r.FormValue("duration_end"),
		PersonnelCount:       personnelCount,
		ExecutiveSummary:     r.FormValue("executive_summary"),
		KeyEvents:            r.FormValue("key_events"),
		WhatWentWell:         r.FormValue("what_went_well"),
		NeedsImprovement:     r.FormValue("needs_improvement"),
		LessonsLearned:       r.FormValue("lessons_learned"),
		Recommendations:      r.FormValue("recommendations"),
		CommandersAssessment: r.FormValue("commanders_assessment"),
		PreparedBy:           r.FormValue("prepared_by"),
		ReviewedBy:           r.FormValue("reviewed_by"),
		Status:               r.FormValue("status"),
		Attachments:          []models.Attachment{},
	}

	// Handle file uploads if S3 client is configured
	if s3Client != nil {
		files := r.MultipartForm.File["attachments"]
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				showCreateForm(w, fmt.Errorf("error opening file: %w", err), "")
				return
			}
			defer file.Close()

			// Generate S3 key
			s3Key := fmt.Sprintf("aars/%s/attachments/%s", aarID, fileHeader.Filename)

			// Upload to S3
			ctx := context.Background()
			if err := s3Client.UploadFile(ctx, s3Key, file, fileHeader.Header.Get("Content-Type")); err != nil {
				showCreateForm(w, fmt.Errorf("error uploading file to S3: %w", err), "")
				return
			}

			// Create attachment record
			attachment := models.Attachment{
				ID:          fmt.Sprintf("att-%d", time.Now().UnixNano()),
				AARID:       aarID,
				Filename:    fileHeader.Filename,
				S3Key:       s3Key,
				FileSize:    fileHeader.Size,
				ContentType: fileHeader.Header.Get("Content-Type"),
				UploadedAt:  now,
			}
			aar.Attachments = append(aar.Attachments, attachment)
		}
	}

	// Save AAR to storage
	if err := store.Create(aar); err != nil {
		showCreateForm(w, fmt.Errorf("error saving AAR: %w", err), "")
		return
	}

	// Show success message
	showCreateForm(w, nil, aarID)
}
