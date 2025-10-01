package api

// UploadResponse is returned after a successful upload.
type UploadResponse struct {
	ID string `json:"id"`
}

// ErrorResponse is a simple error envelope.
type ErrorResponse struct {
	Error string `json:"error"`
}
