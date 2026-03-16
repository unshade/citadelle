package repositories

import "gorm.io/gorm"

type Database struct {
	ServerNodes ServerNodesRepo
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{
		ServerNodes: NewServerNodesRepo(db),
	}
}
