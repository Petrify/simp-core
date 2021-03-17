package service

import (
	"database/sql"
	"log"
)

type Service interface {
	Start() error
	Stop()
	Running() bool
	Status() Status
	giveMessage(*message)
	abstract() *AbstractService

	//getters & setters
	Name() string
	ID() int64
}

type AbstractService struct {
	Logger log.Logger
	name   string
	db     *sql.DB //all services have a database connection
	msgIn  chan *message
	uid    int64
}

func (s *AbstractService) giveMessage(m *message) {
	s.msgIn <- m
}

func (s *AbstractService) ID() int64 {
	return s.uid
}

func (s *AbstractService) Name() string {
	return s.name
}

func (s *AbstractService) abstract() *AbstractService {
	return s
}
