package service

import (
	"fmt"
	"log"
)

type ServiceCtor func(id int64, name string, logger *log.Logger) Service

type sType struct {
	name string
	ex   bool
	ctor ServiceCtor
}

var (
	types    map[string]sType
	services map[int64]Service
	sysServ  *SysService
	sCount   int
)

func NewSType(name string, cTor ServiceCtor, ex bool) error {
	if _, ok := types[name]; !ok {
		typ := sType{name, ex, cTor}
		types[name] = typ
		return nil
	} else {
		return DuplicateTypeError(fmt.Errorf("Service Type with Name [%s] already exists", name))
	}
}

func NewService(typ string, id int64, name string) (*Service, error) {
	return newService(typ, id, name)
}

func newService(typ string, id int64, name string) (*Service, error) {
	//create new service
	sTyp, ok := types[typ]
	if !ok {
		return nil, InvalidTypeError(fmt.Errorf("Service type [%s] does not exist", typ))
	}
	newServ := sTyp.ctor(id, name, sysServ.Log)

	//assign typ
	newServ.abstract().typ = typ

	model, err := sysServ.qService(id)
	if err != nil {
		return nil, err
	} else if model == nil {
		err = newServ.Setup()
		if err != nil {
			return nil, err
		}
		err = sysServ.dbNewService(id, name, typ, false)
		if err != nil {
			return nil, err
		}
	}

	//register service
	err = registerService(newServ)
	if err != nil {
		return nil, fmt.Errorf("could not register service [%d] %s :%e", newServ.ID(), newServ.Name(), err)
	}
	return &newServ, nil
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
