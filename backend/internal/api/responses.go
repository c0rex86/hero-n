package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// APIResponse - базовый ответ API
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta - метаданные ответа
type Meta struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id,omitempty"`
	Version   string `json:"version"`
}

// APIError - структура ошибки API
type APIError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   string                 `json:"details,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// ErrorResponse - ответ с ошибкой
type ErrorResponse struct {
	Success bool     `json:"success"`
	Error   APIError `json:"error"`
}

// ValidationError - ошибка валидации
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// Error codes
const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
	ErrCodeForbidden     = "FORBIDDEN"
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeConflict      = "CONFLICT"
	ErrCodeInternal      = "INTERNAL_ERROR"
	ErrCodeNotImplemented = "NOT_IMPLEMENTED"
)

// Common error messages
const (
	ErrMsgValidationFailed = "Некорректные данные запроса"
	ErrMsgUnauthorized     = "Не авторизован"
	ErrMsgForbidden        = "Недостаточно прав"
	ErrMsgNotFound         = "Ресурс не найден"
	ErrMsgConflict         = "Конфликт данных"
	ErrMsgInternal         = "Внутренняя ошибка сервера"
	ErrMsgNotImplemented   = "Функционал не реализован"
)

// SendSuccess - отправляет успешный ответ
func (s *Server) SendSuccess(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Success: true,
		Data:    data,
	}

	s.writeJSON(w, http.StatusOK, response)
}

// SendCreated - отправляет ответ "создано"
func (s *Server) SendCreated(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Success: true,
		Data:    data,
	}

	s.writeJSON(w, http.StatusCreated, response)
}

// SendError - отправляет ошибку
func (s *Server) SendError(w http.ResponseWriter, status int, code, message string) {
	response := ErrorResponse{
		Success: false,
		Error: APIError{
			Code:      code,
			Message:   message,
			Timestamp: jsonTimeNow(),
		},
	}

	s.writeJSON(w, status, response)
}

// SendErrorWithDetails - отправляет ошибку с деталями
func (s *Server) SendErrorWithDetails(w http.ResponseWriter, status int, code, message, details string) {
	response := ErrorResponse{
		Success: false,
		Error: APIError{
			Code:      code,
			Message:   message,
			Details:   details,
			Timestamp: jsonTimeNow(),
		},
	}

	s.writeJSON(w, status, response)
}

// SendValidationError - отправляет ошибку валидации
func (s *Server) SendValidationError(w http.ResponseWriter, errors []ValidationError) {
	fields := make(map[string]interface{})
	for _, err := range errors {
		fields[err.Field] = map[string]interface{}{
			"message": err.Message,
			"value":   err.Value,
		}
	}

	response := ErrorResponse{
		Success: false,
		Error: APIError{
			Code:      ErrCodeValidation,
			Message:   ErrMsgValidationFailed,
			Fields:    fields,
			Timestamp: jsonTimeNow(),
		},
	}

	s.writeJSON(w, http.StatusBadRequest, response)
}

// SendNotImplemented - отправляет ошибку "не реализовано"
func (s *Server) SendNotImplemented(w http.ResponseWriter, feature string) {
	s.SendErrorWithDetails(w, http.StatusNotImplemented,
		ErrCodeNotImplemented,
		ErrMsgNotImplemented,
		"Функция: "+feature)
}

// SendUnauthorized - отправляет ошибку авторизации
func (s *Server) SendUnauthorized(w http.ResponseWriter) {
	s.SendError(w, http.StatusUnauthorized, ErrCodeUnauthorized, ErrMsgUnauthorized)
}

// SendForbidden - отправляет ошибку доступа
func (s *Server) SendForbidden(w http.ResponseWriter) {
	s.SendError(w, http.StatusForbidden, ErrCodeForbidden, ErrMsgForbidden)
}

// SendNotFound - отправляет ошибку "не найдено"
func (s *Server) SendNotFound(w http.ResponseWriter, resource string) {
	s.SendErrorWithDetails(w, http.StatusNotFound, ErrCodeNotFound, ErrMsgNotFound, "Ресурс: "+resource)
}

// SendInternalError - отправляет внутреннюю ошибку
func (s *Server) SendInternalError(w http.ResponseWriter, err error) {
	s.logger.Printf("Internal error: %v", err)
	s.SendError(w, http.StatusInternalServerError, ErrCodeInternal, ErrMsgInternal)
}

// jsonTimeNow - возвращает текущее время в формате JSON
func jsonTimeNow() string {
	return jsonTime(time.Now())
}

// jsonTime - конвертирует время в JSON формат
func jsonTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05Z07:00")
}
