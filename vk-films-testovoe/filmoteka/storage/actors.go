package sqlite

import (
	"database/sql"
	"log/slog"
)

type Actor struct {
	ActorId   int		`json:"id,omitempty"`
	Name      string	`json:"name"`
	Gender    string	`json:"gender"`
	BirthDate string	`json:"birthdate"`
	Films     []string	`json:"films"`
}

// //Актёры
//
//	app.GetAllActors(log, storage, w, r)
func GetAllActorsFromStorage(s *Storage, log *slog.Logger) ([]Actor, error) {
	rows, err := s.db.Query("SELECT ActorId,Name,Gender,BirthDate FROM Actors")

	log.Info("starting to get actors from storage")

	res := []Actor{}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		actor := Actor{}

		err := rows.Scan(&actor.ActorId, &actor.Name, &actor.Gender, &actor.BirthDate)
		if err != nil {
			return nil, err
		}

		actor.Films, err = filmsForActor(s, actor.ActorId)
		if err != nil {
			return nil, err
		}

		res = append(res, actor)
	}

	log.Info("get all actors from storage")

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// для получения списка фильмов актёра
func filmsForActor(s *Storage, actorID int) ([]string, error) {
	var films []string

	rows, err := s.db.Query(`
		SELECT Films.Title
		FROM Films
		JOIN ActorFilm ON Films.FilmId = ActorFilm.FilmId
		WHERE ActorFilm.ActorId = :id
	`, 
	sql.Named("id", actorID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var filmTitle string
		if err := rows.Scan(&filmTitle); err != nil {
			return nil, err
		}
		films = append(films, filmTitle)
	}

	return films, nil
}

// app.PostActor(log, storage, w, r)
func PostActorToStorage(s *Storage, actor Actor) error {
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

	result, err := s.db.Exec("INSERT INTO Actors (Name, Gender, BirthDate) VALUES (:Name, :Gender, :BirthDate)",
	sql.Named("Name", actor.Name),
	sql.Named("Gender", actor.Gender),
	sql.Named("BirthDate", actor.BirthDate))
	if err != nil {
		return err
	}
	actorID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	for _, movies := range actor.Films {
		var filmID int64
		err = tx.QueryRow("SELECT FilmId FROM Films WHERE Title = :Title", sql.Named("Title", movies)).Scan(&filmID)
		if err != nil {
			if err == sql.ErrNoRows {
				// Фильм не найден, добавляем новый фильм
				result, err := tx.Exec("INSERT INTO Films (Title, Description,Rating,ReleaseDate) VALUES (:Title,:Description,:Rating,:ReleaseDate)",
				 sql.Named("Title", movies),
				 sql.Named("Description", ""),
				 sql.Named("Rating", 0),
				 sql.Named("ReleaseDate", ""),
				)
				if err != nil {
					return err
				}
				filmID, err = result.LastInsertId()
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		_, err = tx.Exec("INSERT INTO ActorFilm (ActorId, FilmId) VALUES (:ActorId, :FilmId)", sql.Named("ActorId", actorID), sql.Named("FilmId", filmID))
		if err != nil {
			return err
		}
	}

    return nil
}

// app.GetOneActor(log, storage, w, r)
func GetOneActorFromStorage(s *Storage, id int, log *slog.Logger) (Actor, error) {
	row := s.db.QueryRow("SELECT ActorId,Name,Gender,BirthDate FROM Actors WHERE ActorId = :id", sql.Named("id", id))

	log.Info("starting get actor from storage")

	var actor Actor

	err := row.Scan(&actor.ActorId, &actor.Name, &actor.Gender, &actor.BirthDate)
	if err != nil {
		return Actor{}, err
	}

	actor.Films, err = filmsForActor(s, actor.ActorId)
	if err != nil {
		return Actor{}, err
	}

	log.Info("get actor from storage successfully")

	return actor, nil
}

// 		app.PutOneActor(log, storage, w, r)
func UpdateActor(s *Storage, actor Actor) error {
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

    _, err = tx.Exec("UPDATE Actors SET Name=:Name, Gender=:Gender, BirthDate=:BirthDate WHERE ActorId = :id",
        sql.Named("Name", actor.Name),
        sql.Named("Gender", actor.Gender),
        sql.Named("BirthDate", actor.BirthDate),
        sql.Named("id", actor.ActorId))
    if err != nil {
        return err
    }

    _, err = tx.Exec("DELETE FROM ActorFilm WHERE ActorId = :id", sql.Named("id", actor.ActorId))
    if err != nil {
        return err
    }

    for _, movie := range actor.Films {
        var filmID int64
        err = tx.QueryRow("SELECT FilmId FROM Films WHERE Title = :Title", sql.Named("Title", movie)).Scan(&filmID)
        if err != nil {
            if err == sql.ErrNoRows {
				result, err := tx.Exec("INSERT INTO Films (Title, Description,Rating,ReleaseDate) VALUES (:Title,:Description,:Rating,:ReleaseDate)",
				sql.Named("Title", movie),
				sql.Named("Description", ""),
				sql.Named("Rating", 0),
				sql.Named("ReleaseDate", ""),
			   	)
                if err != nil {
                    return err
                }
                filmID, err = result.LastInsertId()
                if err != nil {
                    return err
                }
            } else {
                return err
            }
        }
        _, err = tx.Exec("INSERT INTO ActorFilm (ActorId, FilmId) VALUES (:ActorId, :FilmId)",
            sql.Named("ActorId", actor.ActorId),
            sql.Named("FilmId", filmID))
        if err != nil {
            return err
        }
    }

    return nil
}

// 		app.DeleteOneActor(log, storage, w, r)
func DeleteActor(s *Storage, actorID int) error {
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

    _, err = tx.Exec("DELETE FROM ActorFilm WHERE ActorId=:id", sql.Named("id", actorID))
    if err != nil {
        return err
    }

    _, err = tx.Exec("DELETE FROM Actors WHERE ActorId=:id", sql.Named("id", actorID))
    if err != nil {
        return err
    }

    return nil
}


