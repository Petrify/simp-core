package service

import (
	"database/sql"
	"log"
)

func init() {
}

func StartSysService(db *sql.DB, logger *log.Logger) {

	// Check if the System service already exists
	//if it does, no op
	if _, ok := services[0]; ok {
		return
	}
	s := SysService{
		*NewAbstractService("System", 0, db, logger),
	}
	registerService(&s)

}

type SysService struct {
	AbstractService
}

func (s *SysService) Setup() error {
	return nil
}

func (s *SysService) Init() error {

	sList, err := qStartupServices(s.DB)
	if err != nil {
		s.Log.Println("Error Getting Startup Services")
		return err //TODO:
	}

	for _, ms := range sList {

		//create a new service
		err := newService(ms.typ, ms.id, ms.name, s.DB)
		if err != nil {
			s.Log.Print(err)
			continue
		}
	}
	return nil
}

func (s *SysService) Start() error {

	for id, serv := range services {

		//don't start self
		if id == 0 {
			continue
		}

		//Start Service
		err := serv.Start()
		if err != nil {
			s.Log.Printf("Could not start service [%d] %s \n%e", serv.ID(), serv.Name(), err)
			continue
		}
	}

	return nil
}

//
func (s *SysService) Stop() {
	for id, serv := range services {

		//don't stop self
		if id == 0 {
			continue
		}

		//Start Service
		err := serv.Start()
		if err != nil {
			s.Log.Printf("Could not start service [%d] %s \n%e", serv.ID(), serv.Name(), err)
			continue
		}
	}
}

func (s *SysService) Running() bool {
	return true
}
