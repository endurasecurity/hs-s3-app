package models

import "time"

// AAR represents an After Action Report with Defense/Military standard fields
type AAR struct {
	// Identification & Classification
	ID               string    `json:"id"`                // Auto-generated (format: AAR-YYYYMMDD-####)
	Classification   string    `json:"classification"`    // UNCLASSIFIED, CUI, CONFIDENTIAL, SECRET
	OperationName    string    `json:"operation_name"`    // e.g., "Operation Steel Guardian"
	DTG              string    `json:"dtg"`               // Date-Time Group (e.g., "071430ZNOV25")
	UnitDesignation  string    `json:"unit_designation"`  // e.g., "1st Battalion, 75th Ranger Regiment"

	// Operational Details
	MissionType      string    `json:"mission_type"`      // Training Exercise, Combat Operations, etc.
	Location         string    `json:"location"`          // Area of Operations (AO)
	DurationStart    string    `json:"duration_start"`    // Start DTG
	DurationEnd      string    `json:"duration_end"`      // End DTG
	PersonnelCount   int       `json:"personnel_count"`   // Number of personnel involved

	// AAR Content (Narrative Sections)
	ExecutiveSummary   string `json:"executive_summary"`    // Brief overview
	KeyEvents          string `json:"key_events"`           // Chronological significant events
	WhatWentWell       string `json:"what_went_well"`       // Successes
	NeedsImprovement   string `json:"needs_improvement"`    // Areas requiring improvement
	LessonsLearned     string `json:"lessons_learned"`      // Key takeaways
	Recommendations    string `json:"recommendations"`      // Future recommendations
	CommandersAssessment string `json:"commanders_assessment"` // Commander's input (optional)

	// Administrative
	PreparedBy       string    `json:"prepared_by"`       // Name and rank
	ReviewedBy       string    `json:"reviewed_by"`       // Name and rank (optional)
	SubmittedDate    time.Time `json:"submitted_date"`    // Auto-timestamp
	Status           string    `json:"status"`            // Draft, Submitted, Under Review, Approved, Archived

	// Attachments (stored in S3)
	Attachments      []Attachment `json:"attachments"`

	// Timestamps
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Attachment represents a file stored in S3 linked to an AAR
type Attachment struct {
	ID           string    `json:"id"`
	AARID        string    `json:"aar_id"`
	Filename     string    `json:"filename"`
	S3Key        string    `json:"s3_key"`         // Full S3 key path
	FileSize     int64     `json:"file_size"`      // Size in bytes
	ContentType  string    `json:"content_type"`   // MIME type
	UploadedAt   time.Time `json:"uploaded_at"`
}

// Mission type constants
const (
	MissionTypeTraining      = "Training Exercise"
	MissionTypeCombat        = "Combat Operations"
	MissionTypeHumanitarian  = "Humanitarian Assistance"
	MissionTypeSecurity      = "Security Cooperation"
	MissionTypeOther         = "Other"
)

// Classification level constants
const (
	ClassificationUnclassified = "UNCLASSIFIED"
	ClassificationCUI          = "CUI"
	ClassificationConfidential = "CONFIDENTIAL"
	ClassificationSecret       = "SECRET"
)

// Status constants
const (
	StatusDraft       = "Draft"
	StatusSubmitted   = "Submitted"
	StatusUnderReview = "Under Review"
	StatusApproved    = "Approved"
	StatusArchived    = "Archived"
)
