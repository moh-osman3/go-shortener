package managers

import (
	"context"
	"net/http"
)

type UrlManager interface {
	CreateUrlHandleFunc(w http.ResponseWriter, r *http.Request)
	DeleteUrlHandleFunc(w http.ResponseWriter, r *http.Request)
	GetUrlHandleFunc(w http.ResponseWriter, r *http.Request)
	Start(ctx context.Context) error
	End()
}
