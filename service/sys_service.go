package service

import (
	"log"
)

func init() {
}

func StartSysService(logger *log.Logger, sysName string) {

	// Check if the System service already exists
	//if it does, no op
	if _, ok := services[0]; ok {
		return
	}
	s := SysService{
		sysName:         sysName,
		AbstractService: *NewAbstractService("System", 0, logger),
	}
	sysServ = &s
	sysServ.typ = "system"
	err := registerService(&s)
	if err != nil {
		panic(err)
	}

	err = s.Start()
	if err != nil {
		panic(err)
	}

}

type SysService struct {
	sysName string //name of system as given by server_config.yml
	AbstractService
}

func (s *SysService) Setup() error {
	return BuildSchema(s, "sys_schema.sql")
}

func (s *SysService) Init() error {

	// In case of first-time setup, build the system schema
	ok, err := SchemaExists(s)
	if err != nil {
		return err
	} else if !ok {
		BuildSchema(s, "sys_schema.sql")
	}

	sList, err := s.qStartupServices()
	if err != nil {
		s.Log.Println("Error Getting Startup Services: ", err)
		return err //TODO:
	}

	for _, ms := range sList {

		//create a new service
		err := NewService(ms.typ, ms.id, ms.name)
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

		s.Log.Printf("[%d] %s Started Successfully", serv.ID(), serv.Name())
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
