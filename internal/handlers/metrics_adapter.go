package handlers

import (
	"context"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/email"
)

type EmailHealthChecker interface {
	GetStatus(ctx context.Context) (*email.HealthStatus, error)
}

type metricsDataSource struct {
	adminService AdminDashboardService
	emailChecker EmailHealthChecker
	database     db.Database
}

func NewMetricsDataSource(adminSvc AdminDashboardService, emailChk EmailHealthChecker, database db.Database) MetricsDataSource {
	return &metricsDataSource{
		adminService: adminSvc,
		emailChecker: emailChk,
		database:     database,
	}
}

func (s *metricsDataSource) GetAdminStats(ctx context.Context) (*AdminMetricsStats, error) {
	stats, err := s.adminService.GetAdminStats(ctx)
	if err != nil {
		return nil, err
	}
	return &AdminMetricsStats{
		TotalUsers:   stats.TotalUsers,
		TotalEvents:  stats.TotalEvents,
		TotalInvites: stats.TotalInvites,
	}, nil
}

func (s *metricsDataSource) GetEmailQueueStatus(ctx context.Context) (*EmailQueueMetrics, error) {
	status, err := s.emailChecker.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &EmailQueueMetrics{
		QueueSize:    status.QueueSize,
		SendingCount: status.SendingCount,
		FailedCount:  status.FailedCount,
		Healthy:      status.Healthy,
	}, nil
}

func (s *metricsDataSource) GetDBStats() (*DBPoolMetrics, error) {
	return NewDBPoolMetricsFromSQLDB(s.database.DB()), nil
}

var _ MetricsDataSource = (*metricsDataSource)(nil)
