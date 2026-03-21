package httpx

import (
	"net/http"
)

type StatusRecorder struct {
	http.ResponseWriter
	status int
}

func NewStatusRecorder(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *StatusRecorder) Status() int {
	return r.status
}
