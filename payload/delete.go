package payload

type Delete struct {
	ID       string `json:"id"`
	DeleteBy string `json:"delete_by"`
}
