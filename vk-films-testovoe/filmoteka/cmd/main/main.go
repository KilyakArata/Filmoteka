package main

import (
	slog "log/slog"
	"net/http"
	"os"

	"vk-testovoe/filmoteka/config"
	"vk-testovoe/filmoteka/storage"
	"vk-testovoe/filmoteka/app"

	_ "modernc.org/sqlite"
)

func main() {
	cfg := config.MustLoad()

	log:=settupLogger()

	log.Info("starting vk-films-testovoe")
	log.Debug("debug messages are enabled")

	storage, err := sqlite.New(cfg.StoragePath, log)
	if err != nil {
		log.Error("failed to init storage")
		os.Exit(1)
	}

	log.Info("storage connected")

	r := http.NewServeMux()

	//Актёры
	r.HandleFunc("/actors", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.GetAllActors(log, storage, w, r)
		case http.MethodPost:
			app.PostActor(log, storage, w, r)
		}
	})
	r.HandleFunc("/actors/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.GetOneActor(log, storage, w, r)
		case http.MethodPut:
			app.PutOneActor(log, storage, w, r)
		case http.MethodDelete:
			app.DeleteOneActor(log, storage, w, r)
		}
	})

	//Фильмы
	r.HandleFunc("/films", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.GetAllFilms(log, storage, w, r)
		case http.MethodPost:
			app.PostFilm(log, storage, w, r)
		}
	})

	r.HandleFunc("/films/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.GetOneFilm(log, storage, w, r)
		case http.MethodPut:
			app.PutOneFilm(log, storage, w, r)
		case http.MethodDelete:
			app.DeleteOneFilm(log, storage, w, r)
		}
	})

	srv := &http.Server{
		Addr:         cfg.Address,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
		Handler:      r,
	}

	log.Info("starting vk-films-testovoe on server", slog.String("server", cfg.Address))

	err = srv.ListenAndServe()
	if err != nil {
		log.Error("failed to start server:", err)
		return
	}

	log.Info("server started")
}

func settupLogger() *slog.Logger {
	var log *slog.Logger
	log=slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	return log
}
