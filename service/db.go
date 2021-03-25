package service

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"

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

	err := simpsql.ExecScriptSchema(s.DB, "new_service", Schema(s), id, name, typ, startup)
	if err != nil {
		sysServ.Log.Println("error while saving new service to database", err)
		return err
	}

	return nil
}

//Finds search by service ID
func (s *SysService) qService(id int64) (*modelService, error) {
	rows, err := simpsql.QueryScriptSchema(s.DB, "get_service", Schema(s), id)
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

	rows, err := simpsql.QueryScriptSchema(s.DB, "get_startup_services", Schema(s))
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

func buildSchema(schema string, fPath string, db *sql.DB) error {

	//get path of package that called the public buildSchema functions
	_, b, _, _ := runtime.Caller(2)
	basepath := filepath.Dir(b)

	//read script from file and split into statements
	fRead, err := ioutil.ReadFile(filepath.Join(basepath, fPath))
	if err != nil {
		return err
	}
	script := string(fRead)
	stmts := strings.Split(script, ";")

	//begin new transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	//execute each statement, error and cancel transaction on error
	for _, stmt := range stmts {
		//skip accidental empty statements
		if stmt == "" {
			sysServ.Log.Println("Found empty SQL statement")
			continue
		}
		stmt = fmt.Sprintf(stmt, schema)
		sysServ.Log.Print("Executing SQL: ", stmt)
		_, err := tx.Exec(stmt)
		if err != nil {
			sysServ.Log.Println("Ran into an error: ", err)
			tx.Rollback()
			return err
		}

	}

	return tx.Commit()
}
