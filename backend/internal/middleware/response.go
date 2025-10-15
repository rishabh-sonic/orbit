package middleware

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Ok(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, APIResponse{Data: data})
}

func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, APIResponse{Data: data})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func BadRequest(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusBadRequest, APIResponse{Error: msg})
}

func Unauthorized(w http.ResponseWriter, msg string) {
	if msg == "" {
		msg = "unauthorized"
	}
	JSON(w, http.StatusUnauthorized, APIResponse{Error: msg})
}

func Forbidden(w http.ResponseWriter, msg string) {
	if msg == "" {
		msg = "forbidden"
	}
	JSON(w, http.StatusForbidden, APIResponse{Error: msg})
}

func NotFound(w http.ResponseWriter, msg string) {
	if msg == "" {
		msg = "not found"
	}
	JSON(w, http.StatusNotFound, APIResponse{Error: msg})
}

func Conflict(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusConflict, APIResponse{Error: msg})
}

func InternalError(w http.ResponseWriter, err error) {
	JSON(w, http.StatusInternalServerError, APIResponse{Error: "internal server error"})
}
