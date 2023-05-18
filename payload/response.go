package payload

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}
