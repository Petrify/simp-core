package service

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
)

type modelService struct {
	name string
	id   int64
	typ  string
}

//checks if a Database Already exists
func dbExists(s Service) (bool, error) {
	r, err := s.abstract().DB.Query("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", Schema(s))
	if err != nil {
		return false, err
	}
	return r.Next(), nil
}

//Finds all services marked with startup
func (s *SysService) qStartupServices() (lst []modelService, err error) {
	stmt, err := s.DB.Prepare(fmt.Sprintf("SELECT servicename, serviceid, servicetype FROM %s.service WHERE startupservice = 1", Schema(s)))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	//scan each element returned to a Service and append it to lst
	for rows.Next() {
		s := modelService{}
		err = rows.Scan(&s.name, &s.id, &s.typ)
		if err != nil {
			return nil, err
		}
		lst = append(lst, s)
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
	stmts := strings.Split(script, ";\n")

	//begin new transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	//execute each statement, error and cancel transaction on error
	for _, stmt := range stmts {
		stmt = fmt.Sprintf(stmt, schema)
		_, err = tx.Exec(stmt)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
