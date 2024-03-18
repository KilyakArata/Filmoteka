package sqlite

import (
	"database/sql"
	"log/slog"
)

type Film struct {
	FilmId      int			`json:"id,omitempty"`
	Title       string		`json:"title"`
	Description string		`json:"description"`
	Rating      int			`json:"rating"`
	ReleaseDate	string		`json:"releaseDate"`
	Actors      []string	`json:"actors"`
}

// //Фильмы
// 		app.GetAllFilms(log, storage, w, r)
func GetAllFilmsFromStorage(s *Storage, log *slog.Logger) ([]Film, error) {
	rows, err := s.db.Query("SELECT FilmId,Title,Description,Rating,ReleaseDate FROM Films")

	log.Info("starting to get all films from storage")

	res := []Film{}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		film := Film{}

		err := rows.Scan(&film.FilmId, &film.Title, &film.Description, &film.Rating, &film.ReleaseDate)
		if err != nil {
			return nil, err
		}

		film.Actors, err = actorsForFilm(s, film.FilmId)
		if err != nil {
			return nil, err
		}

		res = append(res, film)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	log.Info(" get all films from storage successfully")

	return res, nil
}

// поиск актёров одного фильма
func actorsForFilm(s *Storage, filmID int) ([]string, error) {
	var actors []string

	rows, err := s.db.Query(`
		SELECT Actors.Name
		FROM Actors
		JOIN ActorFilm ON Actors.ActorId = ActorFilm.ActorId
		WHERE ActorFilm.FilmId = :id
	`,
	sql.Named("id",filmID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var actorName string
		if err := rows.Scan(&actorName); err != nil {
			return nil, err
		}
		actors = append(actors, actorName)
	}

	return actors, nil
}

// 		app.PostFilm(log, storage, w, r)
func PostFilmToStorage(s *Storage, film Film) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	result, err := s.db.Exec("INSERT INTO Films (Title, Description, Rating, ReleaseDate) VALUES (:Title, :Description, :Rating, :ReleaseDate)",
	sql.Named("Title", film.Title),
	sql.Named("Description", film.Description),
	sql.Named("Rating", film.Rating),
	sql.Named("ReleaseDate", film.ReleaseDate))
	if err != nil {
		return err
	}
	filmID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	for _, act := range film.Actors {
		var actorID int64
		err = tx.QueryRow("SELECT ActorId FROM Actors WHERE Name = :Name", sql.Named("Name", act)).Scan(&actorID)
		if err != nil {
			if err == sql.ErrNoRows {
				result, err := tx.Exec("INSERT INTO Actors (Name,Gender,BirthDate) VALUES (:Name,:Gender,:BirthDate)",
				 sql.Named("Name", act),
				 sql.Named("Gender", ""),
				 sql.Named("BirthDate", ""),
				)
				if err != nil {
					return err
				}
				actorID, err = result.LastInsertId()
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		_, err = tx.Exec("INSERT INTO ActorFilm (ActorId, FilmId) VALUES (:ActorId, FilmId)", sql.Named("ActorID", actorID), sql.Named("FilmID", filmID))
		if err != nil {
			return err
		}
	}

    return nil
}

// 		app.GetOneFilm(log, storage, w, r)
func GetOneFilmFromStorage(s *Storage, id int) (Film, error) {
	row := s.db.QueryRow("SELECT FilmId,Title,Description,Rating,ReleaseDate FROM Films WHERE FilmId = :id", sql.Named("id", id))

	var film Film

	err := row.Scan(&film.FilmId, &film.Title,&film.Description,&film.Rating, &film.ReleaseDate)
	if err != nil {
		return Film{}, err
	}

	film.Actors, err = filmsForActor(s, film.FilmId)
	if err != nil {
		return Film{}, err
	}

	return film, nil
}
// 		app.PutOneFilm(log, storage, w, r)
func UpdateFilm(s *Storage, film Film) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer func() {
        if err != nil {
            tx.Rollback()
            return
        }
        err = tx.Commit()
    }()


    _, err = tx.Exec("UPDATE Films SET Title=:Title, Description=:Description, Rating=:Rating, ReleaseDate=:ReleaseDate WHERE FilmId = :id",
        sql.Named("Title", film.Title),
        sql.Named("Description", film.Description),
        sql.Named("Rating", film.Rating),
        sql.Named("ReleaseDate", film.ReleaseDate),
        sql.Named("id", film.FilmId))
    if err != nil {
        return err
    }


    _, err = tx.Exec("DELETE FROM ActorFilm WHERE FilmId = :id", sql.Named("id", film.FilmId))
    if err != nil {
        return err
    }


    for _, actor := range film.Actors {
        var actorID int64

        err = tx.QueryRow("SELECT ActorId FROM Actors WHERE Name = :Name", sql.Named("Name", actor)).Scan(&actorID)
        if err != nil {
            if err == sql.ErrNoRows {

				result, err := tx.Exec("INSERT INTO Actors (Name,Gender,BirthDate) VALUES (:Name,:Gender,:BirthDate)",
				 sql.Named("Name", actor),
				 sql.Named("Gender", ""),
				 sql.Named("BirthDate", ""),
				)
                if err != nil {
                    return err
                }
                actorID, err = result.LastInsertId()
                if err != nil {
                    return err
                }
            } else {
                return err
            }
        }

        _, err = tx.Exec("INSERT INTO ActorFilm (ActorId, FilmId) VALUES (:ActorId, :FilmId)",
            sql.Named("ActorId", actorID),
            sql.Named("FilmId", film.FilmId))
        if err != nil {
            return err
        }
    }

    return nil
}
// 		app.DeleteOneFilm(log, storage, w, r)
func DeleteFilm(s *Storage, filmID int) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer func() {
        if err != nil {
            tx.Rollback()
            return
        }
        err = tx.Commit()
    }()


    _, err = tx.Exec("DELETE FROM ActorFilm WHERE FilmId=:id", sql.Named("id", filmID))
    if err != nil {
        return err
    }


    _, err = tx.Exec("DELETE FROM Films WHERE FilmId=:id", sql.Named("id", filmID))
    if err != nil {
        return err
    }

    return nil
}
