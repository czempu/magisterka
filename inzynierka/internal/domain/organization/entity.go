package organization

type OrgUnit struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	ParentID *int   `json:"parent_id"`
}

type OrgMember struct {
	UserID    int    `json:"user_id"`
	OrgUnitID int    `json:"org_unit_id"`
	Title     string `json:"title"`
	IsPrimary bool   `json:"is_primary"`
}

type OrgManager struct {
	OrgUnitID int  `json:"org_unit_id"`
	UserID    int  `json:"user_id"`
	IsDefault bool `json:"is_default"`
}

