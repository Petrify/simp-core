package service

import (
	"database/sql"
	"fmt"
)

type ServiceCtor func(int, string, sql.DB) Service

var (
	cTors    map[string]ServiceCtor
	services map[int64]Service
	idByName map[string]int64
	sCount   int
)

func NewServiceType(typeName string, cTor ServiceCtor) {
	
}

func RegisterService(s Service) error {
	if _, ok := services[s.ID()]; ok {
		return DuplicateIDError(fmt.Errorf("service with ID [%d] already exists", s.ID()))
	} else if _, ok := idByName[s.Name()]; ok {
		return DuplicateNameError(fmt.Errorf("service with Name [%s] already exists", s.Name()))
	} else {
		services[s.ID()] = s
		idByName[s.Name()] = s.ID()
		return nil
	}
}

func ServiceById(id int64) (Service, error) {

	if s, ok := services[id]; ok {
		return s, nil
	} else {
		return nil, NoSuchServiceError(fmt.Errorf("no service with ID [%d]", id))
	}

}

func GetService(name string) (Service, error) {

	if id, ok := idByName[name]; ok {
		return ServiceById(id)
	} else {
		return nil, NoSuchServiceError(fmt.Errorf("no service with name [%s]", name))
	}
}

func RenameService(curr string, name string) error {
	s, e := GetService(curr)
	if e != nil {
		return e
	} 

	if _, ok := idByName[name]; ok {
		return DuplicateNameError(fmt.Errorf("service with Name [%s] already exists", name))
	}
	
	s.abstract().name = name
	idByName[name] = s.ID()
	delete(idByName, curr)

	return nil
}

type DuplicateIDError error

type DuplicateNameError error

type DuplicateTypeError error

type NoSuchServiceError error
