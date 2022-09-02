package handlers

import (
	"encoding/json"
	"github.com/GeorgeShibanin/InternWB/internal/storage"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

func (h *HTTPHandler) HandleGetData(rw http.ResponseWriter, r *http.Request) {
	key := strings.Trim(r.URL.Path, "/")
	data, err := h.storage.GetData(r.Context(), storage.Id(key))
	if err != nil || data.OrderUID == "" {
		http.NotFound(rw, r)
		return
	}
	//http.Redirect(rw, r, string(url), http.StatusPermanentRedirect)
	response := PutResponseData{
		Model: data,
	}
	rawResponse, err := json.Marshal(response)
	if err != nil {
		err = errors.Wrap(err, "can't marshal response")
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	_, err = rw.Write(rawResponse)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}
