package response

import (
	"encoding/json"
	"net/http"
)

// SuccessResponse writes a success response with a given status code and data
func SuccessResponse(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ErrorResponse writes an error response with a given status code and message
func ErrorResponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
