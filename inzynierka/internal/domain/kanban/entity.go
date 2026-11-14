package kanban

type Ticket struct {
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	Desc      string   `json:"description"`
	Type      string   `json:"type"`
	Status    string   `json:"status"`
	Assignees []string `json:"assignees"`
}

type CalendarEntry struct {
	Person  string `json:"person"`
	DateStr string `json:"date_str"`
	Status  int    `json:"status"`
}
