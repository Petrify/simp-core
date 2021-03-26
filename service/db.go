package service

import (
	"database/sql"
	"fmt"

	simpsql "github.com/Petrify/simp-core/sql"
)

type modelService struct {
	name string
	id   int64
	typ  string
	v    int
}

//checks if a Database Already exists
func SchemaExists(s Service) (bool, error) {
	return simpsql.SchemaExists(sysServ.DB, Schema(s))
}

func SubschemaExists(s Service, subname string) (bool, error) {
	return simpsql.SchemaExists(sysServ.DB, Subschema(s, subname))
}

func (s *SysService) dbNewService(id int64, name string, typ string, startup bool) error {

	err := simpsql.ExecScriptSchema(s.DB, "sys_new_service.sql", Schema(s), id, name, typ, startup)
	if err != nil {
		sysServ.Log.Println("error while saving new service to database", err)
		return err
	}

	return nil
}

//Finds search by service ID
func (s *SysService) qService(id int64) (*modelService, error) {
	rows, err := simpsql.QueryScriptSchema(s.DB, "sys_get_service.sql", Schema(s), id)
	if err != nil {
		return nil, err
	}

	if rows.Next() {
		ms := modelService{}
		err = rows.Scan(&ms.name, &ms.id, &ms.typ, &ms.v)
		if err != nil {
			return nil, err
		}
		return &ms, nil
	}

	return nil, nil
}

//Finds all services marked with startup
func (s *SysService) qStartupServices() (lst []modelService, err error) {

	rows, err := simpsql.QueryScriptSchema(s.DB, "sys_get_startup_services.sql", Schema(s))
	if err != nil {
		return nil, err
	}

	//scan each element returned to a Service and append it to lst
	for rows.Next() {
		ms := modelService{}
		err = rows.Scan(&ms.name, &ms.id, &ms.typ, &ms.v)
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

func Subschema(s Service, subname string) string {
	return fmt.Sprintf("%s_%d_%s", sysServ.sysName, s.ID(), subname)
}

func BuildSchema(s Service, fPath string) error {
	schema := Schema(s)
	return buildSchema(schema, fPath, s.abstract().DB)
}

func BuildSubschema(s Service, name string, fPath string) error {
	schema := Subschema(s, name)
	return buildSchema(schema, fPath, s.abstract().DB)
}

func buildSchema(schema string, scriptPath string, db *sql.DB) error {

	return simpsql.ExecScriptSchema(db, scriptPath, schema)

}
