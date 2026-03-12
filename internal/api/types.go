package api

type Project struct {
	Self                 string            `json:"self"`
	TeamID               *string           `json:"team_id"`
	Title                string            `json:"title"`
	TitleChain           []string          `json:"title_chain"`
	Color                string            `json:"color"`
	ProductivityScore    *float64          `json:"productivity_score"`
	IsArchived           bool              `json:"is_archived"`
	Notes                *string           `json:"notes"`
	Children             []Project         `json:"children,omitempty"`
	Parent               *ProjectRef       `json:"parent"`
	DefaultBillingStatus string            `json:"default_billing_status"`
	CustomFields         map[string]string `json:"custom_fields"`
}

type ProjectRef struct {
	Self string `json:"self"`
}

type TimeEntry struct {
	Self          string            `json:"self"`
	StartDate     string            `json:"start_date"`
	EndDate       string            `json:"end_date"`
	Duration      float64           `json:"duration"`
	Project       *Project          `json:"project"`
	Title         string            `json:"title"`
	Notes         string            `json:"notes"`
	IsRunning     bool              `json:"is_running"`
	CreatorID     string            `json:"creator_id"`
	CreatorName   string            `json:"creator_name"`
	BillingStatus string            `json:"billing_status"`
	CustomFields  map[string]string `json:"custom_fields"`
}

type ReportRow struct {
	Duration float64  `json:"duration"`
	Project  *Project `json:"project"`
	Title    string   `json:"title,omitempty"`
	Notes    string   `json:"notes,omitempty"`
	User     string   `json:"user,omitempty"`
	Timespan string   `json:"timespan,omitempty"`
}

type DataResponse struct {
	Data    any    `json:"data"`
	Message string `json:"message,omitempty"`
}
