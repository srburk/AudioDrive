package router

import (
	. "audiodrive/internal/controller"
	"net/http"
)

func NewHandler(controllers ...Controller) http.Handler {
	mux := http.NewServeMux()

	for _, c := range controllers {
		c.RegisterRoutes(mux)
	}

	return mux
}
