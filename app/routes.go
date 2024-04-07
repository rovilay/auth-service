package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rovilay/auth-service/handlers"
	"github.com/rs/cors"
)

func (a *App) loadRoutes() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var res struct {
			Message string
		}

		res.Message = "Welcome to Auth service"

		msg, err := json.Marshal(res)
		if err != nil {
			fmt.Println("failed to marshall ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(msg)
	})

	router.Route("/inventory", a.loadUserRoutes)

	// CORS configuration
	corsRouter := cors.Default().Handler(router)

	a.router = corsRouter
}

func (a *App) loadUserRoutes(router chi.Router) {
	h := handlers.NewUserHandler(a.repo, a.log)

	router.Post("/signup", h.Signup)
	router.Post("/login", h.Login)

	router.Group(func(r chi.Router) {
		r.Use(h.MiddlewareAuth)
		r.Get("/user/{id}", h.GetUser)
	})
}
