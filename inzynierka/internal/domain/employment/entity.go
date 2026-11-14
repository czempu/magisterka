package employment


type EmploymentTerm struct {
	ID                int       `json:"id"`
	UserID            int       `json:"user_id"`
	EffectiveFrom     string `json:"effective_from"`
	EffectiveTo       *string `json:"effective_to,omitempty"`
	FTEPercent        float64   `json:"fte_percent"`
	Position          string    `json:"position"`
}


type EmploymentTermAdd struct {
	UserID            int       `json:"user_id"`
	EffectiveFrom     string `json:"effective_from"`
	EffectiveTo       *string `json:"effective_to,omitempty"`
	FTEPercent        float64   `json:"fte_percent"`
	Position          string    `json:"position"`
}
