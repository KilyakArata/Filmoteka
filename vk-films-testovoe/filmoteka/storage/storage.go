package sqlite

import (
	"database/sql"
	"log/slog"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string, log *slog.Logger) (*Storage, error) {
	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		log.Error("failed to open storage:", err)
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Actors (
        ActorId INTEGER PRIMARY KEY,
        Name TEXT UNIQUE,
        Gender TEXT,
        BirthDate TEXT
    )`)
	if err != nil {
		log.Error("failed to create table Actors: %w", err)
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Films (
        FilmId INTEGER PRIMARY KEY,
        Title TEXT UNIQUE,
        Description TEXT,
		ReleaseDate TEXT,
        Rating INTEGER
    )`)
	if err != nil {
		log.Error("failed to create table Films: %w", err)
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ActorFilm (
        ActorId INTEGER,
        FilmId INTEGER,
        PRIMARY KEY (ActorId, FilmId),
        FOREIGN KEY (ActorId) REFERENCES Actors (ActorId),
        FOREIGN KEY (FilmId) REFERENCES Films (FilmId)
    )`)
	if err != nil {
		log.Error("failed to create table ActorFilm: %w", err)
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Users (
        Login TEXT PRIMARY KEY,
        Password TEXT
    )`)
	if err != nil {
		log.Error("failed to create table Users: %w", err)
		return nil, err
	}


	rows, err := db.Query("SELECT COUNT(*) FROM Users WHERE Login IN ('Admin', 'User')")
	if err != nil {
		log.Error("failed to query Users table:", err)
		return nil, err
	}
	

	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Error("failed to scan row:", err)
			return nil, err
		}
	}
	rows.Close()

	if count==0 {
		_, err = db.Exec(`INSERT INTO Users (Login, Password) VALUES ('Admin', 'c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f')`)
		if err != nil {
			log.Error("failed to insert role Admin in table Users: ", err)
			return nil, err
		}

		_, err = db.Exec(`INSERT INTO Users (Login, Password) VALUES ('User', 'b512d97e7cbf97c273e4db073bbb547aa65a84589227f8f3d9e4a72b9372a24d')`)
		if err != nil {
			log.Error("failed to insert role User in table Users: ", err)
			return nil, err
		}
	}

	
 
	return &Storage{db: db}, nil
}
