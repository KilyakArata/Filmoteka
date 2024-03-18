package sqlite

func GetUsers(s *Storage) (map[string]string, error) {
	rows, err := s.db.Query("SELECT Login, Password FROM Users")
	a := make(map[string]string)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var login string
		var password string
		err := rows.Scan(&login, &password)
		if err != nil {
			return nil, err
		}

		a[login] = password
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return a, nil
}
