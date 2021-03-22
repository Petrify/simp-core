package service

import (
	"database/sql"
	"fmt"
)

var ServerName string

type modelService struct {
	name string
	id   int64
	typ  string
}

//Finds all services marked with startup
func qStartupServices(db *sql.DB) (lst []modelService, err error) {
	stmt, err := db.Prepare("SELECT servicename, serviceid, servicetype FROM simp.service WHERE startupservice = 1")
	if err != nil {
		return nil, err
	}

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
	return fmt.Sprintf("%s_%s_$d", ServerName, s.abstract().typ, s.ID())
}
