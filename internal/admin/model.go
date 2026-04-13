package admin

type ReviewTask struct {
	ID             string
	SkillVersionID string
	NamespaceID    string
	Status         string
}

type ReviewDecisionInput struct {
	TaskID     string `json:"taskId"`
	Decision   string `json:"decision"`
	Comment    string `json:"comment"`
	ReviewerID string
}
