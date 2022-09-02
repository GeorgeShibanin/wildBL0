package handlers

import (
	"github.com/GeorgeShibanin/InternWB/internal/storage"
)

type HTTPHandler struct {
	storage  storage.Storage
	inMemory map[string]storage.Orders
}

func NewHTTPHandler(storage1 storage.Storage) *HTTPHandler {
	return &HTTPHandler{
		storage: storage1,
	}
}

type PutRequestData struct {
	Model storage.Orders
}

type PutResponseKey struct {
	Id string `json:"OrderUid"`
}

type PutResponseData struct {
	Model storage.Orders
}
