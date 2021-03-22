package service

import (
	"database/sql"
	"log"
)

type Service interface {
	//Setup should perform first time setup routines for the Service such as creating database tables
	Setup() error

	//Init runs once at the Registration of the Service. acts like a secondary constructor
	Init() error
	Start() error
	Stop()
	//Status() Status

	//implemented by AbstractService
	giveMessage(*message)
	abstract() *AbstractService
	Name() string
	ID() int64
}

type AbstractService struct {
	name  string
	id    int64
	typ   string
	DB    *sql.DB //all services have a database connection
	msgIn chan *message
	Log   *log.Logger
}

func NewAbstractService(name string, id int64, db *sql.DB, logger *log.Logger) *AbstractService {
	s := AbstractService{
		name:  name,
		id:    id,
		typ:   "",
		DB:    db,
		msgIn: make(chan *message),
		Log:   logger,
	}

	return &s
}

func (s *AbstractService) giveMessage(m *message) {
	s.msgIn <- m
}

func (s *AbstractService) ID() int64 {
	return s.id
}

func (s *AbstractService) Name() string {
	return s.name
}

func (s *AbstractService) abstract() *AbstractService {
	return s
}
