package backup

type RestoreResponse struct {
	Inserted *[]string `json:"inserted"`
	Existing *[]string `json:"existing"`
}
