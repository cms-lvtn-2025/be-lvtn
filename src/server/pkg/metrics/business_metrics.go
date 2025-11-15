package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Domain-specific business metrics

	// Thesis Management Metrics
	ThesisSubmissionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "thesis_submissions_total",
			Help: "Total number of thesis submissions",
		},
		[]string{"academic_year", "department", "status"},
	)

	ThesisReviewsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "thesis_reviews_total",
			Help: "Total number of thesis reviews",
		},
		[]string{"reviewer_type", "status", "department"},
	)

	ThesisApprovalTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "thesis_approval_duration_days",
			Help:    "Time taken for thesis approval in days",
			Buckets: []float64{1, 3, 7, 14, 30, 60, 90},
		},
		[]string{"review_stage", "department"},
	)

	ActiveThesesByStage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_theses_by_stage",
			Help: "Number of active theses by stage",
		},
		[]string{"stage", "department"},
	)

	// Council Management Metrics
	CouncilMeetingsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "council_meetings_total",
			Help: "Total number of council meetings",
		},
		[]string{"meeting_type", "department"},
	)

	CouncilDecisionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "council_decisions_total",
			Help: "Total number of council decisions",
		},
		[]string{"decision_type", "outcome"},
	)

	CouncilMembersActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "council_members_active",
			Help: "Number of active council members",
		},
		[]string{"role", "department"},
	)

	// User Activity Metrics
	UserLoginFrequency = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_login_frequency_total",
			Help: "User login frequency by role and time",
		},
		[]string{"role", "department", "time_period"},
	)

	UserSessionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_session_duration_minutes",
			Help:    "User session duration in minutes",
			Buckets: []float64{5, 15, 30, 60, 120, 240, 480},
		},
		[]string{"role", "activity_type"},
	)

	ActiveUsersGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of active users by role",
		},
		[]string{"role", "department"},
	)

	// File Management Metrics
	FileUploadsByType = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "file_uploads_by_type_total",
			Help: "Total file uploads by type",
		},
		[]string{"file_type", "category", "department"},
	)

	FileStorageUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "file_storage_usage_bytes",
			Help: "File storage usage in bytes",
		},
		[]string{"category", "department"},
	)

	FileProcessingTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "file_processing_duration_seconds",
			Help:    "File processing duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60},
		},
		[]string{"operation", "file_type"},
	)

	// Academic Performance Metrics
	StudentGradesDistribution = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "student_grades_distribution",
			Help:    "Distribution of student grades",
			Buckets: []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		[]string{"subject", "semester", "department"},
	)

	DeadlineMissed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "deadlines_missed_total",
			Help: "Total number of missed deadlines",
		},
		[]string{"deadline_type", "severity", "department"},
	)

	// System Performance Metrics
	NotificationsSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_sent_total",
			Help: "Total notifications sent",
		},
		[]string{"type", "channel", "status"},
	)

	EmailDeliveryTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "email_delivery_duration_seconds",
			Help:    "Email delivery duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"template_type"},
	)
)

// Business metrics recording functions

// Thesis Metrics
func RecordThesisSubmission(academicYear, department, status string) {
	ThesisSubmissionsTotal.WithLabelValues(academicYear, department, status).Inc()
}

func RecordThesisReview(reviewerType, status, department string) {
	ThesisReviewsTotal.WithLabelValues(reviewerType, status, department).Inc()
}

func RecordThesisApprovalTime(reviewStage, department string, days float64) {
	ThesisApprovalTime.WithLabelValues(reviewStage, department).Observe(days)
}

func SetActiveThesesByStage(stage, department string, count float64) {
	ActiveThesesByStage.WithLabelValues(stage, department).Set(count)
}

// Council Metrics
func RecordCouncilMeeting(meetingType, department string) {
	CouncilMeetingsTotal.WithLabelValues(meetingType, department).Inc()
}

func RecordCouncilDecision(decisionType, outcome string) {
	CouncilDecisionsTotal.WithLabelValues(decisionType, outcome).Inc()
}

func SetCouncilMembersActive(role, department string, count float64) {
	CouncilMembersActive.WithLabelValues(role, department).Set(count)
}

// User Activity Metrics
func RecordUserLogin(role, department, timePeriod string) {
	UserLoginFrequency.WithLabelValues(role, department, timePeriod).Inc()
}

func RecordUserSessionDuration(role, activityType string, duration time.Duration) {
	UserSessionDuration.WithLabelValues(role, activityType).Observe(duration.Minutes())
}

func SetActiveUsers(role, department string, count float64) {
	ActiveUsersGauge.WithLabelValues(role, department).Set(count)
}

// File Metrics
func RecordFileUpload(fileType, category, department string) {
	FileUploadsByType.WithLabelValues(fileType, category, department).Inc()
}

func SetFileStorageUsage(category, department string, bytes float64) {
	FileStorageUsage.WithLabelValues(category, department).Set(bytes)
}

func RecordFileProcessingTime(operation, fileType string, duration time.Duration) {
	FileProcessingTime.WithLabelValues(operation, fileType).Observe(duration.Seconds())
}

// Academic Metrics
func RecordStudentGrade(subject, semester, department string, grade float64) {
	StudentGradesDistribution.WithLabelValues(subject, semester, department).Observe(grade)
}

func RecordDeadlineMissed(deadlineType, severity, department string) {
	DeadlineMissed.WithLabelValues(deadlineType, severity, department).Inc()
}

// System Metrics
func RecordNotificationSent(notificationType, channel, status string) {
	NotificationsSent.WithLabelValues(notificationType, channel, status).Inc()
}

func RecordEmailDeliveryTime(templateType string, duration time.Duration) {
	EmailDeliveryTime.WithLabelValues(templateType).Observe(duration.Seconds())
}
