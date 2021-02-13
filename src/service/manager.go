package service

import "database/sql"

type Manager struct {
	db       sql.DB
	services []Service
}

func NewManager(db sql.DB) (err error, m Manager) {

}
