package dto

type ImportResult struct {
	SuccessfulCount int      `json:"successful_count"`
	SuccessfulItems []string `json:"successful_items"`
	FailedCount     int      `json:"failed_count"`
	FailedItems     []string `json:"failed_items"`
}
