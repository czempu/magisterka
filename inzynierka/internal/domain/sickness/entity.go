package sickness

// Sickness – pojedyncze zgłoszenie chorobowe (L4)
type Sickness struct {
	ID                     int    `json:"id"`
	EmployeeID             int    `json:"employee_id"`
	FromDate               string `json:"from_date"`   // YYYY-MM-DD
	ToDate                 string `json:"to_date"`     // YYYY-MM-DD
	LegalBasis             string `json:"legal_basis"` // np. "Ustawa zasiłkowa art. ..."
	CertificateNumber      string `json:"certificate_number"`
	PercentCoverage        int    `json:"percent_coverage"`         // 80, 100 itd.
	IsOccupationalAccident bool   `json:"is_occupational_accident"` // true/false
	Status                 string `json:"status"`                   // pending/approved/...
	ReportedAt             string `json:"reported_at"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
}

