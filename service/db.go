package service

import (
	"fmt"

	simpsql "github.com/Petrify/simp-core/sql"
)

type modelService struct {
	name    string
	id      int64
	typ     string
	version int
}

//checks if a Database Already exists
func SchemaExists(s Service) (bool, error) {
	return simpsql.SchemaExists(Schema(s))
}

func (s *SysService) dbNewService(id int64, name string, typ string, startup bool) error {

	tx, err := simpsql.UsingSchema(Schema(s))
	if err != nil {
		return err
	}
	defer tx.Commit()

	_, err = tx.Exec(
		`INSERT INTO service
		(serviceid,
		servicename,
		servicetype,
		startupservice)
		VALUES (?,?,?,?)`,
		id, name, typ, startup)

	if err != nil {
		tx.Rollback()
		sysServ.Log.Printf("An Error while saving new service `%s` to database: \n%e", name, err)
		return err
	}

	return nil
}

//Finds search by service ID
func (s *SysService) qService(id int64) (*modelService, error) {

	tx, err := simpsql.UsingSchema(Schema(s))
	if err != nil {
		return nil, err
	}
	defer tx.Commit()

	rows, err := tx.Query(
		`SELECT servicename, serviceid, servicetype, version FROM service WHERE serviceid = ?`,
		id)

	if err != nil {
		sysServ.Log.Printf("An Error while getting service with id `%d`: \n%e", id, err)
		return nil, err
	}

	if rows.Next() {
		ms := modelService{}
		err = rows.Scan(&ms.name, &ms.id, &ms.typ, &ms.version)
		if err != nil {
			return nil, err
		}
		return &ms, nil
	}

	return nil, nil
}

//Finds all services marked with startup
func (s *SysService) qStartupServices() (lst []modelService, err error) {

	tx, err := simpsql.UsingSchema(Schema(s))
	if err != nil {
		return nil, err
	}
	defer tx.Commit()

	rows, err := tx.Query("SELECT servicename, serviceid, servicetype, version FROM `service` WHERE startupservice = 1")
	if err != nil {
		sysServ.Log.Printf("An Error while getting startup services: \n%e", err)
		return nil, err
	}

	//scan each element returned to a Service and append it to lst
	for rows.Next() {
		ms := modelService{}
		err = rows.Scan(&ms.name, &ms.id, &ms.typ, &ms.version)
		if err != nil {
			return nil, err
		}
		lst = append(lst, ms)
	}

	return
}

//returns string name of the schema associated with a service
func Schema(s Service) string {
	return fmt.Sprintf("%s_%s_%d", sysServ.sysName, s.abstract().typ, s.ID())
}

func BuildSchema(s Service, scriptPath string) error {
	schema := Schema(s)

	err := simpsql.MakeSchema(schema)
	if err != nil {
		return err
	}

	tx, err := simpsql.UsingSchema(schema)
	if err != nil {
		return err
	}

	err = simpsql.ExecScript(tx, scriptPath)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error executing script `%s`: %e", scriptPath, err)
	}

	tx.Commit()
	return nil
}
