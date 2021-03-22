package service

import (
	"database/sql"
	"log"
)

func init() {
}

func StartSysService(db *sql.DB, logger *log.Logger, sysName string) {

	// Check if the System service already exists
	//if it does, no op
	if _, ok := services[0]; ok {
		return
	}
	s := SysService{
		sysName:         sysName,
		AbstractService: *NewAbstractService("System", 0, db, logger),
	}
	sysServ = &s
	sysServ.typ = "system"
	registerService(&s)

}

type SysService struct {
	sysName string //name of system as given by server_config.yml
	AbstractService
}

func (s *SysService) Setup() error {
	return BuildSchema(s, "sys_schema.sql")
}

func (s *SysService) Init() error {

	ok, err := dbExists(s)
	if err != nil {
		return err
	} else if !ok {
		BuildSchema(s, "sys_schema.sql")
	}

	sList, err := s.qStartupServices()
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
