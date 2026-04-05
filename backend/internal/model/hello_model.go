package model

// HelloResponse is returned by the /api/hello endpoint.
type HelloResponse struct {
	Message string `json:"message"`
	Time    string `json:"time"`
}
