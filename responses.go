package main

import (
	"encoding/json"
	"net/http"
)

// APIError represents a structured error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// APIResponse represents a standardized API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// Error codes for consistent frontend handling
const (
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeSessionExpired     = "SESSION_EXPIRED"
	ErrCodeRateLimited        = "RATE_LIMITED"
	ErrCodeLinkNotFound       = "LINK_NOT_FOUND"
	ErrCodeInvalidLinkData    = "INVALID_LINK_DATA"
	ErrCodeSaveFailed         = "SAVE_FAILED"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeMethodNotAllowed   = "METHOD_NOT_ALLOWED"
	ErrCodeInvalidRequest     = "INVALID_REQUEST"
	ErrCodeAvatarNotFound     = "AVATAR_NOT_FOUND"
	ErrCodeInvalidFile        = "INVALID_FILE"
	ErrCodeConfigReloadFailed = "CONFIG_RELOAD_FAILED"
)

// writeJSONError writes a standardized JSON error response
func writeJSONError(w http.ResponseWriter, status int, code, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// writeJSONSuccess writes a standardized JSON success response
func writeJSONSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

// writeJSONSuccessWithStatus writes a standardized JSON success response with custom status
func writeJSONSuccessWithStatus(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

// Helper functions for common error scenarios
func writeUnauthorizedError(w http.ResponseWriter, message string) {
	writeJSONError(w, http.StatusUnauthorized, ErrCodeUnauthorized, message, "Please log in to access this resource")
}

func writeMethodNotAllowedError(w http.ResponseWriter, method string) {
	writeJSONError(w, http.StatusMethodNotAllowed, ErrCodeMethodNotAllowed,
		"Method not allowed",
		"The "+method+" method is not supported for this endpoint")
}

func writeInvalidRequestError(w http.ResponseWriter, message, details string) {
	writeJSONError(w, http.StatusBadRequest, ErrCodeInvalidRequest, message, details)
}

func writeInternalServerError(w http.ResponseWriter, message, details string) {
	writeJSONError(w, http.StatusInternalServerError, ErrCodeSaveFailed, message, details)
}

func writeNotFoundError(w http.ResponseWriter, resource string) {
	writeJSONError(w, http.StatusNotFound, ErrCodeLinkNotFound,
		resource+" not found",
		"The requested "+resource+" does not exist")
}
