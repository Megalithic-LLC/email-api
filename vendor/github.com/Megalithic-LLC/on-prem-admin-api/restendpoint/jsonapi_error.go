package restendpoint

type JsonApiSource struct {
	Pointer   string `json:"pointer"`
	Parameter string `json:"parameter,omitempty"`
}

type JsonApiError struct {
	Status string        `json:"status,omitempty"`
	Title  string        `json:"title,omitempty"`
	Detail string        `json:"detail,omitempty"`
	Source JsonApiSource `json:"source,omitempty"`
}
