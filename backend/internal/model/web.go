package model

// WebResponse is the generic envelope returned from every HTTP endpoint.
type WebResponse[T any] struct {
	Data T `json:"data"`
}
