package repositories

import "gorm.io/gorm"

type Database struct {
	ServerNodes ServerNodesRepo
	Users       UsersRepo
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{
		ServerNodes: NewServerNodesRepo(db),
		Users:       NewUsersRepo(db),
	}
}
