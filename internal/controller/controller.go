package controller

import (
	"net/http"
)

type Controller interface {
	RegisterRoutes(mux *http.ServeMux)
}
