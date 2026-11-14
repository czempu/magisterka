package schedule

import (
	"context"
)

// ReportGenerator defines the interface for generating exportable reports (PDF, CSV, etc.)
type ReportGenerator interface {
	GenerateMonthlyPDF(ctx context.Context, report MonthlyReport, userIDStr, month string) (string, error)
}
