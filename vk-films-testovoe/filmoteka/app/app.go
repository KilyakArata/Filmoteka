package app

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	
	"vk-testovoe/filmoteka/storage"
	"vk-testovoe/filmoteka/verify"
)

const(
	layout = "02.01.2006"
)

func GetAllActors(log *slog.Logger, s *sqlite.Storage,w http.ResponseWriter, r *http.Request){
	user, pass, ok := r.BasicAuth()
	if !ok {
		log.Error("unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ok=verify.User(user, pass, log, verify.ReadPermission, s)
	if !ok {
		log.Error("wrong role")
		http.Error(w, "wrong role", http.StatusUnauthorized)
		return
	}
	
	sliceOfActors,err:=sqlite.GetAllActorsFromStorage(s, log)
	if err != nil {
		log.Error("no list of actors: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(sliceOfActors)
	if err != nil {
		log.Error("cant json.marshal actors: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
	log.Info("get actors successfully")
}

func PostActor(log *slog.Logger, s *sqlite.Storage,w http.ResponseWriter, r *http.Request){
	user, pass, ok := r.BasicAuth()
	if !ok {
		log.Error("unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	ok=verify.User(user, pass, log, verify.WritePermission, s)
	if !ok {
		log.Error("wrong role")
		http.Error(w, "wrong role", http.StatusUnauthorized)
		return
	}
	log.Info("authorization was successful")

	var actor sqlite.Actor
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Error("wrong input:",err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &actor); err != nil {
		log.Error("wrong unmarshal inputted:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if actor.Gender!="male" && actor.Gender!="female"{
		log.Error("wrong gender")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "wrong gender")
		return
	}

	_, err = time.Parse(layout, actor.BirthDate)
	if err != nil {
		log.Error("wrong BirthDate")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err=sqlite.PostActorToStorage(s,actor)
	if err != nil {
		log.Error("error to post actor to storage:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, "Actor posted")
}


func GetOneActor(log *slog.Logger, s *sqlite.Storage,w http.ResponseWriter, r *http.Request){
	user, pass, ok := r.BasicAuth()
	if !ok {
		log.Error("unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ok=verify.User(user, pass, log, verify.ReadPermission, s)
	if !ok {
		log.Error("wrong role")
		http.Error(w, "wrong role", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		log.Error("invalid URL path")
		http.Error(w, "invalid URL path", http.StatusBadRequest)
		return
	}
	actorID, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Error("invalid actor ID:", err)
		http.Error(w, "invalid actor ID", http.StatusBadRequest)
		return
	}
	
	actor,err:=sqlite.GetOneActorFromStorage(s, actorID, log)
	if err != nil {
		log.Error("no actor:", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(actor)
	if err != nil {
		log.Error("cant json.marshal actor:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
	log.Info("get actor successfully")
}


func PutOneActor(log *slog.Logger, s *sqlite.Storage,w http.ResponseWriter, r *http.Request){
	user, pass, ok := r.BasicAuth()
	if !ok {
		log.Error("unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	ok=verify.User(user, pass, log, verify.WritePermission, s)
	if !ok {
		log.Error("wrong role")
		http.Error(w, "wrong role", http.StatusUnauthorized)
		return
	}
	log.Info("authorization was successful")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		log.Error("invalid URL path")
		http.Error(w, "invalid URL path", http.StatusBadRequest)
		return
	}
	actorID, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Error("invalid actor ID:", err)
		http.Error(w, "invalid actor ID", http.StatusBadRequest)
		return
	}

	var actor sqlite.Actor
	var buf bytes.Buffer

	actor.ActorId=actorID

	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		log.Error("wrong input:",err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &actor); err != nil {
		log.Error("wrong unmarshal inputted:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if (actor.Gender!="male") && (actor.Gender!="female"){
		log.Error("wrong gender")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "wrong gender")
		return
	}

	_, err = time.Parse(layout, actor.BirthDate)
	if err != nil {
		log.Error("wrong BirthDate")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err=sqlite.UpdateActor(s,actor)
	if err != nil {
		log.Error("error to post actor to storage:", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, "Actor updated")
}


func DeleteOneActor(log *slog.Logger, s *sqlite.Storage,w http.ResponseWriter, r *http.Request){
	user, pass, ok := r.BasicAuth()
	if !ok {
		log.Error("unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ok=verify.User(user, pass, log, verify.WritePermission, s)
	if !ok {
		log.Error("wrong role")
		http.Error(w, "wrong role", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		log.Error("invalid URL path")
		http.Error(w, "invalid URL path", http.StatusBadRequest)
		return
	}
	actorID, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Error("invalid actor ID:", err)
		http.Error(w, "invalid actor ID", http.StatusBadRequest)
		return
	}
	
	err=sqlite.DeleteActor(s, actorID)
	if err != nil {
		log.Error("no actor:", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	io.WriteString(w, "Actor deleted")
	log.Info("actor deleted successfully")
}

func GetAllFilms(log *slog.Logger, s *sqlite.Storage,w http.ResponseWriter, r *http.Request){
	user, pass, ok := r.BasicAuth()
	if !ok {
		log.Error("unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ok=verify.User(user, pass, log, verify.ReadPermission, s)
	if !ok {
		log.Error("wrong role")
		http.Error(w, "wrong role", http.StatusUnauthorized)
		return
	}
	
	sliceOfFilms,err:=sqlite.GetAllFilmsFromStorage(s, log)
	if err != nil {
		log.Error("no list of films: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(sliceOfFilms)
	if err != nil {
		log.Error("cant json.marshal films: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
	log.Info("get films successfully")
}

func PostFilm(log *slog.Logger, s *sqlite.Storage,w http.ResponseWriter, r *http.Request){
	user, pass, ok := r.BasicAuth()
	if !ok {
		log.Error("unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	ok=verify.User(user, pass, log, verify.WritePermission, s)
	if !ok {
		log.Error("wrong role")
		http.Error(w, "wrong role", http.StatusUnauthorized)
		return
	}
	log.Info("authorization was successful")

	var film sqlite.Film
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Error("wrong input:",err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &film); err != nil {
		log.Error("wrong unmarshal inputted:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(film.Title)<1 || len(film.Title)>150{
		log.Error("wrong title")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "wrong title")
		return
	}

	if len(film.Description)>1000{
		log.Error("too long description")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "too long description")
		return
	}

	if film.Rating<0||film.Rating>10{
		log.Error("wrong rating")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "wrong rating")
		return
	}

	_, err = time.Parse(layout, film.ReleaseDate)
	if err != nil {
		log.Error("wrong ReleaseDate")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err=sqlite.PostFilmToStorage(s,film)
	if err != nil {
		log.Error("error to post film to storage:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, "Film posted")
}


func GetOneFilm(log *slog.Logger, s *sqlite.Storage,w http.ResponseWriter, r *http.Request){
	user, pass, ok := r.BasicAuth()
	if !ok {
		log.Error("unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ok = verify.User(user, pass, log, verify.ReadPermission, s)
	if !ok {
		log.Error("wrong role")
		http.Error(w, "wrong role", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		log.Error("invalid URL path")
		http.Error(w, "invalid URL path", http.StatusBadRequest)
		return
	}
	filmID, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Error("invalid actor ID:", err)
		http.Error(w, "invalid actor ID", http.StatusBadRequest)
		return
	}

	film, err := sqlite.GetOneFilmFromStorage(s, filmID)
	if err != nil {
		log.Error("no film:", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(film)
	if err != nil {
		log.Error("cant json.marshal film:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
	log.Info("get film successfully")
}


func PutOneFilm(log *slog.Logger, s *sqlite.Storage,w http.ResponseWriter, r *http.Request){
	user, pass, ok := r.BasicAuth()
	if !ok {
		log.Error("unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	ok = verify.User(user, pass, log, verify.WritePermission, s)
	if !ok {
		log.Error("wrong role")
		http.Error(w, "wrong role", http.StatusUnauthorized)
		return
	}
	log.Info("authorization was successful")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		log.Error("invalid URL path")
		http.Error(w, "invalid URL path", http.StatusBadRequest)
		return
	}
	filmID, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Error("invalid actor ID:", err)
		http.Error(w, "invalid actor ID", http.StatusBadRequest)
		return
	}

	var film sqlite.Film
	var buf bytes.Buffer

	film.FilmId=filmID

	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		log.Error("wrong input:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &film); err != nil {
		log.Error("wrong unmarshal inputted:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(film.Title)<1 || len(film.Title)>150{
		log.Error("wrong title")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "wrong title")
		return
	}

	if len(film.Description)>1000{
		log.Error("too long description")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "too long description")
		return
	}

	if film.Rating<0||film.Rating>10{
		log.Error("wrong rating")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "wrong rating")
		return
	}

	_, err = time.Parse(layout, film.ReleaseDate)
	if err != nil {
		log.Error("wrong ReleaseDate")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = time.Parse(layout, film.ReleaseDate)
	if err != nil {
		log.Error("wrong ReleaseDate")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = sqlite.UpdateFilm(s, film)
	if err != nil {
		log.Error("error to update film to storage:", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, "Film updated")
}

func DeleteOneFilm(log *slog.Logger, s *sqlite.Storage,w http.ResponseWriter, r *http.Request){
	user, pass, ok := r.BasicAuth()
	if !ok {
		log.Error("unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ok = verify.User(user, pass, log, verify.WritePermission, s)
	if !ok {
		log.Error("wrong role")
		http.Error(w, "wrong role", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		log.Error("invalid URL path")
		http.Error(w, "invalid URL path", http.StatusBadRequest)
		return
	}
	filmID, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Error("invalid actor ID:", err)
		http.Error(w, "invalid actor ID", http.StatusBadRequest)
		return
	}

	err = sqlite.DeleteFilm(s, filmID)
	if err != nil {
		log.Error("no film:", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	io.WriteString(w, "Film deleted")
	log.Info("film deleted successfully")
}