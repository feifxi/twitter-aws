package apiresponse

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Error struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	RequestID string       `json:"requestId,omitempty"`
	Details   []FieldError `json:"details,omitempty"`
}
