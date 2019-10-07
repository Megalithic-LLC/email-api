package restendpoint

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/docktermj/go-logger/logger"
)

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

func sendBadRequestError(w http.ResponseWriter, err error) {
	sendErrors(w, []JsonApiError{
		JsonApiError{
			Status: fmt.Sprintf("%v", http.StatusBadRequest),
			Title:  "Bad Request",
			Detail: err.Error(),
		},
	})
}

func sendInternalServerError(w http.ResponseWriter) {
	sendErrors(w, []JsonApiError{
		JsonApiError{
			Status: fmt.Sprintf("%v", http.StatusInternalServerError),
			Title:  "Internal Error",
			Detail: "An internal error has occurred",
		},
	})
}

func sendErrors(w http.ResponseWriter, errs []JsonApiError) {
	result := map[string]interface{}{
		"errors": errs,
	}
	w.Header().Set("Content-Type", "application/json")
	status := http.StatusInternalServerError
	if len(errs) > 0 {
		status, _ = strconv.Atoi(errs[0].Status)
	}
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}
