package storage

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"hs-s3-app/models"
)

// MemoryStore provides in-memory storage for AARs
type MemoryStore struct {
	aars map[string]*models.AAR
	mu   sync.RWMutex
}

// NewMemoryStore creates a new in-memory store with pre-populated sample data
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		aars: make(map[string]*models.AAR),
	}
	store.loadSampleData()
	return store
}

// GetAll returns all AARs sorted by submitted date (newest first)
func (s *MemoryStore) GetAll() []*models.AAR {
	s.mu.RLock()
	defer s.mu.RUnlock()

	aars := make([]*models.AAR, 0, len(s.aars))
	for _, aar := range s.aars {
		aars = append(aars, aar)
	}

	// Sort by submitted date, newest first
	sort.Slice(aars, func(i, j int) bool {
		return aars[i].SubmittedDate.After(aars[j].SubmittedDate)
	})

	return aars
}

// GetByID retrieves an AAR by its ID
func (s *MemoryStore) GetByID(id string) (*models.AAR, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	aar, exists := s.aars[id]
	if !exists {
		return nil, fmt.Errorf("AAR not found: %s", id)
	}
	return aar, nil
}

// Create adds a new AAR to the store
func (s *MemoryStore) Create(aar *models.AAR) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.aars[aar.ID]; exists {
		return fmt.Errorf("AAR already exists: %s", aar.ID)
	}

	now := time.Now()
	aar.CreatedAt = now
	aar.UpdatedAt = now
	aar.SubmittedDate = now

	s.aars[aar.ID] = aar
	return nil
}

// Update modifies an existing AAR
func (s *MemoryStore) Update(aar *models.AAR) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.aars[aar.ID]; !exists {
		return fmt.Errorf("AAR not found: %s", aar.ID)
	}

	aar.UpdatedAt = time.Now()
	s.aars[aar.ID] = aar
	return nil
}

// Search filters AARs by various criteria
func (s *MemoryStore) Search(operationName, unit, missionType string) []*models.AAR {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*models.AAR

	for _, aar := range s.aars {
		match := true

		if operationName != "" && !strings.Contains(strings.ToLower(aar.OperationName), strings.ToLower(operationName)) {
			match = false
		}
		if unit != "" && !strings.Contains(strings.ToLower(aar.UnitDesignation), strings.ToLower(unit)) {
			match = false
		}
		if missionType != "" && aar.MissionType != missionType {
			match = false
		}

		if match {
			results = append(results, aar)
		}
	}

	// Sort by submitted date, newest first
	sort.Slice(results, func(i, j int) bool {
		return results[i].SubmittedDate.After(results[j].SubmittedDate)
	})

	return results
}

// GetStats returns statistics about stored AARs
func (s *MemoryStore) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total"] = len(s.aars)

	// Count by mission type
	missionTypes := make(map[string]int)
	for _, aar := range s.aars {
		missionTypes[aar.MissionType]++
	}
	stats["by_mission_type"] = missionTypes

	// Count by status
	statuses := make(map[string]int)
	for _, aar := range s.aars {
		statuses[aar.Status]++
	}
	stats["by_status"] = statuses

	return stats
}

// loadSampleData pre-populates the store with sample AARs for demo purposes
func (s *MemoryStore) loadSampleData() {
	now := time.Now()

	// Sample AAR 1: Operation Enduring Shield
	aar1 := &models.AAR{
		ID:              "AAR-20251005-0001",
		Classification:  models.ClassificationUnclassified,
		OperationName:   "Operation Enduring Shield",
		DTG:             "051200ZOCT25",
		UnitDesignation: "3rd Infantry Division, 2nd Brigade Combat Team",
		MissionType:     models.MissionTypeTraining,
		Location:        "Fort Stewart, GA",
		DurationStart:   "051200ZOCT25",
		DurationEnd:     "051800ZOCT25",
		PersonnelCount:  450,
		ExecutiveSummary: "Battalion-level combined arms training exercise focusing on rapid deployment and sustainment operations. Exercise included live-fire maneuvers, logistical coordination, and interoperability drills with supporting units. All training objectives were met with no significant safety incidents.",
		KeyEvents: "0600: Unit assembly and mission brief\n0800: Movement to training area\n1000: Live-fire exercise commenced\n1400: Tactical maneuver phase\n1700: After action review\n1800: Stand down",
		WhatWentWell: "Communications between units exceeded expectations. Logistical support was timely and effective. All personnel demonstrated proficiency in basic combat tasks. Leadership at the platoon level showed strong tactical decision-making.",
		NeedsImprovement: "Coordination with supporting artillery units needs refinement. Some delays in medical evacuation procedures were noted. Night operations revealed gaps in night vision equipment availability.",
		LessonsLearned: "Pre-coordination with supporting elements is critical for mission success. Additional training on CASEVAC procedures is required. Equipment readiness checks must be more thorough before operations.",
		Recommendations: "Increase frequency of integrated training with artillery and air support. Procure additional night vision devices. Conduct quarterly CASEVAC refresher training for all personnel.",
		CommandersAssessment: "Overall excellent performance by the battalion. Unit is combat-ready and capable of executing assigned missions. Recommend continued emphasis on combined arms integration.",
		PreparedBy:      "CPT John Smith, S3 Operations Officer",
		ReviewedBy:      "LTC Michael Johnson, Battalion Commander",
		Status:          models.StatusApproved,
		SubmittedDate:   now.AddDate(0, 0, -32),
		CreatedAt:       now.AddDate(0, 0, -32),
		UpdatedAt:       now.AddDate(0, 0, -32),
		Attachments: []models.Attachment{
			{
				ID:          "att-001",
				AARID:       "AAR-20251005-0001",
				Filename:    "training_photo_001.jpg",
				S3Key:       "aars/AAR-20251005-0001/attachments/training_photo_001.jpg",
				FileSize:    2048576,
				ContentType: "image/jpeg",
				UploadedAt:  now.AddDate(0, 0, -32),
			},
			{
				ID:          "att-002",
				AARID:       "AAR-20251005-0001",
				Filename:    "sitrep.pdf",
				S3Key:       "aars/AAR-20251005-0001/attachments/sitrep.pdf",
				FileSize:    524288,
				ContentType: "application/pdf",
				UploadedAt:  now.AddDate(0, 0, -32),
			},
			{
				ID:          "att-003",
				AARID:       "AAR-20251005-0001",
				Filename:    "equipment_status.xlsx",
				S3Key:       "aars/AAR-20251005-0001/attachments/equipment_status.xlsx",
				FileSize:    102400,
				ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
				UploadedAt:  now.AddDate(0, 0, -32),
			},
		},
	}

	// Sample AAR 2: Exercise Iron Sentinel
	aar2 := &models.AAR{
		ID:              "AAR-20250920-0002",
		Classification:  models.ClassificationUnclassified,
		OperationName:   "Exercise Iron Sentinel",
		DTG:             "201430ZSEP25",
		UnitDesignation: "Marine Expeditionary Unit 26",
		MissionType:     models.MissionTypeSecurity,
		Location:        "Camp Pendleton, CA / Pacific Ocean",
		DurationStart:   "200600ZSEP25",
		DurationEnd:     "221800ZSEP25",
		PersonnelCount:  2200,
		ExecutiveSummary: "Multi-national training exercise with allied forces practicing amphibious assault operations and humanitarian assistance/disaster relief scenarios. Exercise included naval surface warfare, air-ground integration, and logistics over the shore operations. Participating nations included Japan, Australia, and South Korea.",
		KeyEvents: "Day 1: Embarkation and movement to sea\nDay 2: Amphibious rehearsal\nDay 3: Full-scale amphibious assault with live-fire\nDay 4: HA/DR scenario and medical operations\nDay 5: Debarkation and equipment accountability",
		WhatWentWell: "Excellent coordination with allied forces. Communications interoperability exceeded previous exercises. Aviation support was responsive and effective. Medical staff demonstrated outstanding capability in mass casualty scenarios.",
		NeedsImprovement: "Landing craft scheduling caused minor delays. Some language barriers with partner nations during complex operations. Weather monitoring and contingency planning needs enhancement.",
		LessonsLearned: "Importance of liaison officers embedded with partner forces. Need for redundant communications systems. Value of cultural awareness training prior to multinational exercises.",
		Recommendations: "Increase number of liaison officers for future multinational exercises. Develop standardized visual signals for operations with language barriers. Conduct additional weather-related contingency training.",
		CommandersAssessment: "MEU demonstrated exceptional operational capability and readiness. Integration with allied forces was seamless. Unit is fully prepared for real-world contingency operations.",
		PreparedBy:      "MAJ Sarah Williams, MEU Operations Officer",
		ReviewedBy:      "COL Robert Davis, MEU Commander",
		Status:          models.StatusApproved,
		SubmittedDate:   now.AddDate(0, 0, -47),
		CreatedAt:       now.AddDate(0, 0, -47),
		UpdatedAt:       now.AddDate(0, 0, -47),
		Attachments: []models.Attachment{
			{
				ID:          "att-004",
				AARID:       "AAR-20250920-0002",
				Filename:    "amphibious_ops_photo_001.jpg",
				S3Key:       "aars/AAR-20250920-0002/attachments/amphibious_ops_photo_001.jpg",
				FileSize:    3145728,
				ContentType: "image/jpeg",
				UploadedAt:  now.AddDate(0, 0, -47),
			},
			{
				ID:          "att-005",
				AARID:       "AAR-20250920-0002",
				Filename:    "tactical_map.png",
				S3Key:       "aars/AAR-20250920-0002/attachments/tactical_map.png",
				FileSize:    1572864,
				ContentType: "image/png",
				UploadedAt:  now.AddDate(0, 0, -47),
			},
			{
				ID:          "att-006",
				AARID:       "AAR-20250920-0002",
				Filename:    "exercise_video.mp4",
				S3Key:       "aars/AAR-20250920-0002/attachments/exercise_video.mp4",
				FileSize:    52428800,
				ContentType: "video/mp4",
				UploadedAt:  now.AddDate(0, 0, -47),
			},
			{
				ID:          "att-007",
				AARID:       "AAR-20250920-0002",
				Filename:    "allied_coordination_plan.pdf",
				S3Key:       "aars/AAR-20250920-0002/attachments/allied_coordination_plan.pdf",
				FileSize:    819200,
				ContentType: "application/pdf",
				UploadedAt:  now.AddDate(0, 0, -47),
			},
			{
				ID:          "att-008",
				AARID:       "AAR-20250920-0002",
				Filename:    "medical_ops_photo.jpg",
				S3Key:       "aars/AAR-20250920-0002/attachments/medical_ops_photo.jpg",
				FileSize:    2621440,
				ContentType: "image/jpeg",
				UploadedAt:  now.AddDate(0, 0, -47),
			},
		},
	}

	// Sample AAR 4: Exercise Northern Viking
	aar4 := &models.AAR{
		ID:              "AAR-20250815-0004",
		Classification:  models.ClassificationUnclassified,
		OperationName:   "Exercise Northern Viking",
		DTG:             "151500ZAUG25",
		UnitDesignation: "10th Mountain Division, 1st Brigade",
		MissionType:     models.MissionTypeTraining,
		Location:        "Fort Drum, NY / Adirondack Mountains",
		DurationStart:   "150600ZAUG25",
		DurationEnd:     "171800ZAUG25",
		PersonnelCount:  3500,
		ExecutiveSummary: "Brigade-level cold weather and mountain warfare training exercise. Focused on operations in austere, high-altitude environments with emphasis on cold weather survival, mountain mobility, and logistics in challenging terrain. Exercise prepared unit for potential deployment to arctic regions.",
		KeyEvents: "Day 1: Movement to mountain training area and cold weather acclimatization\nDay 2: Mountain mobility training and technical rope operations\nDay 3: Cold weather survival and bivouac operations\nDay 4: Brigade tactical exercise with opposing force",
		WhatWentWell: "Soldiers demonstrated excellent cold weather discipline. Mountain warfare skills improved significantly. Logistics chain functioned effectively despite challenging terrain. Leadership at all levels adapted well to austere conditions.",
		NeedsImprovement: "Some cold weather equipment shortages were identified. Communications in mountainous terrain remains challenging. Additional medical personnel training for cold weather injuries needed.",
		LessonsLearned: "Importance of proper cold weather equipment maintenance. Need for specialized communications equipment for mountain operations. Value of pre-deployment cold weather training for personnel unfamiliar with arctic conditions.",
		Recommendations: "Procure additional cold weather equipment sets. Invest in mountainous terrain communications solutions. Establish cold weather injury prevention program. Conduct annual cold weather refresher training.",
		CommandersAssessment: "Brigade performed exceptionally well in challenging conditions. Unit is prepared for cold weather and mountain operations. Soldiers demonstrated resilience and adaptability.",
		PreparedBy:      "CPT Emily Rodriguez, Brigade S3 Air",
		ReviewedBy:      "COL James Anderson, Brigade Commander",
		Status:          models.StatusApproved,
		SubmittedDate:   now.AddDate(0, 0, -83),
		CreatedAt:       now.AddDate(0, 0, -83),
		UpdatedAt:       now.AddDate(0, 0, -83),
		Attachments: []models.Attachment{
			{
				ID:          "att-011",
				AARID:       "AAR-20250815-0004",
				Filename:    "mountain_training_photo.jpg",
				S3Key:       "aars/AAR-20250815-0004/attachments/mountain_training_photo.jpg",
				FileSize:    2621440,
				ContentType: "image/jpeg",
				UploadedAt:  now.AddDate(0, 0, -83),
			},
			{
				ID:          "att-012",
				AARID:       "AAR-20250815-0004",
				Filename:    "cold_weather_procedures.pdf",
				S3Key:       "aars/AAR-20250815-0004/attachments/cold_weather_procedures.pdf",
				FileSize:    614400,
				ContentType: "application/pdf",
				UploadedAt:  now.AddDate(0, 0, -83),
			},
		},
	}

	// Add all sample AARs to the store
	s.aars[aar1.ID] = aar1
	s.aars[aar2.ID] = aar2
	s.aars[aar4.ID] = aar4
}
