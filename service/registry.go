package service

import (
	"database/sql"
	"fmt"
)

type serviceCtor func(id int64, name string, db *sql.DB) Service

type sType struct {
	name string
	ex   bool
	ctor serviceCtor
}

var (
	types    map[string]sType
	services map[int64]Service
	idByName map[string]int64
	sCount   int
)

func NewSType(name string, cTor serviceCtor, ex bool) error {
	if _, ok := types[name]; ok {
		typ := sType{name, ex, cTor}
		types[name] = typ
		return nil
	} else {
		return DuplicateTypeError(fmt.Errorf("Service Type with Name [%s] already exists", name))
	}
}

func newService(typ string, id int64, name string, db *sql.DB) error {
	//create new service
	sTyp, ok := types[typ]
	if !ok {
		return InvalidTypeError(fmt.Errorf("Service type [%s] does not exist", typ))
	}
	newServ := sTyp.ctor(id, name, db)

	//register service
	err := registerService(newServ)
	if err != nil {
		return fmt.Errorf("Could not register service [%d] %s \n%e", newServ.ID(), newServ.Name(), err)
	}
	return nil
}

func registerService(s Service) error {
	if _, ok := services[s.ID()]; ok {
		return DuplicateIDError(fmt.Errorf("service with ID [%d] already exists", s.ID()))
	} else {
		err := s.Init()
		if err != nil {
			return InitializationError(err)
		}
		services[s.ID()] = s
		sCount++
		return nil
	}
}

func getService(id int64) (Service, error) {

	if s, ok := services[id]; ok {
		return s, nil
	} else {
		return nil, InvalidServiceError(fmt.Errorf("no service with ID [%d]", id))
	}
}

type DuplicateIDError error

type DuplicateTypeError error

type InitializationError error

type InvalidServiceError error

type InvalidTypeError error
