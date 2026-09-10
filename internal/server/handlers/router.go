package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/scarypuppp/gophkeeper/internal/server/middlewares"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func (h *Handler) GetRouter() http.Handler {
	r := chi.NewRouter()

	mw := middlewares.NewMiddleware(h.config, h.logger)

	r.Use(mw.LogRequest)
	r.Use(mw.LogResponse)

	r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		httpSwagger.Handler()(w, r)
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
		})
		r.Group(func(r chi.Router) {
			r.Use(mw.Auth)
			r.Route("/file", func(r chi.Router) {
				r.Get("/", h.GetFiles)
				r.Post("/upload", h.CreateFile)
				r.Get("/{file_hash}/{file_name}", h.GetFile)
				r.Put("/{file_hash}/{file_name}", h.UpdateFile)
				r.Put("/{file_hash}/{file_name}/metadata", h.UpdateFileMetadata)
				r.Delete("/{file_hash}/{file_name}", h.DeleteFile)
			})
			r.Route("/secret", func(r chi.Router) {
				r.Post("/", h.CreateSecret)
				r.Get("/", h.GetSecrets)
				r.Get("/{secret_name}", h.GetSecret)
				r.Put("/{secret_name}", h.UpdateSecret)
				r.Delete("/{secret_name}", h.DeleteSecret)
			})

			r.Route("/card", func(r chi.Router) {
				r.Post("/", h.CreateCard)
				r.Get("/", h.GetCards)
				r.Get("/{card_name}", h.GetCard)
				r.Put("/{card_name}", h.UpdateCard)
				r.Delete("/{card_name}", h.DeleteCard)
			})

			r.Route("/text", func(r chi.Router) {
				r.Post("/", h.CreateText)
				r.Get("/", h.GetTexts)
				r.Get("/{text_name}", h.GetText)
				r.Put("/{text_name}", h.UpdateText)
				r.Delete("/{text_name}", h.DeleteText)
			})
		})
	})

	return r
}
