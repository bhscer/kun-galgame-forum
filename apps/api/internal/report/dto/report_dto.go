package dto

type SubmitReportRequest struct {
	Reason string `json:"reason" validate:"required,min=10,max=1007"`
	Type   string `json:"type" validate:"required,max=100"`
}
