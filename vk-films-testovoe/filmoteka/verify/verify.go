package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"

	"vk-testovoe/filmoteka/storage"

	_ "modernc.org/sqlite"
)

const (
	ReadPermission  = "read"
	WritePermission = "write"

	AdminRole = "admin"
	UserRole  = "user"
)

var (
	rolePermissions = map[string][]string{
		AdminRole: {ReadPermission, WritePermission},
		UserRole:  {ReadPermission},
	}
)

var (
	userRoles = map[string][]string{
		"User":  {UserRole},
		"Admin": {AdminRole},
	}
)

func User(user, pass string, log *slog.Logger, permission string, s *sqlite.Storage) bool {
	hashedPassword := sha256.Sum256([]byte(pass))
	hashStringPassword := hex.EncodeToString(hashedPassword[:])

	userPassword, err := sqlite.GetUsers(s)
	if err != nil {
		log.Error("cant get users: %v", err)
		return false
	}
	storedPassword, ok := userPassword[user]

	if !ok {
		log.Error("no such user in storage")
		return false
	}

	if hashStringPassword != storedPassword {
		log.Info("wrong password")
		return false
	}

	for _, roles := range userRoles[user] {
		for _, storedPermission := range rolePermissions[roles] {
			if permission == storedPermission {
				log.Info("access is allowed")
				return true
			}
		}
	}

	log.Info("not necessary role")
	return false

}
