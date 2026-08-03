package model

// ReviewScope controls row-level access for reviews and related data.
// OrgID set => shared company workspace; empty => individual user scope.
type ReviewScope struct {
	UserID string
	OrgID  string
}

func (s ReviewScope) IsOrg() bool {
	return s.OrgID != ""
}
