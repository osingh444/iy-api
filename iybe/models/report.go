package models

type Report struct {
	ReviewID       string
	ReviewerID     string
	VendorName     string
	SecondaryIndex string
}

func (report * Report) ResolveReport() (map[string]interface{}, uint) {
	return DeleteReview(report.ReviewID, report.ReviewerID)
}
