package simpsql

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/otiai10/copy"
)

const sqlDelim = ";\n"

var scriptDir = "sql_scripts"
var baseDir string
var absScriptDir string

var DB *sql.DB

func init() {
	exedir, _ := os.Executable()
	baseDir = filepath.Dir(exedir)
	absScriptDir = filepath.Join(baseDir, scriptDir)

	err := os.Mkdir(absScriptDir, os.ModePerm)

	if err == nil {
		cloneSqlScripts()
	}

}

//Clones Base sql_scripts folder from this module to target
func cloneSqlScripts() {
	//get path to this source
	_, b, _, _ := runtime.Caller(0)
	modDir := filepath.Dir(filepath.Dir(b))
	sDir := filepath.Join(modDir, scriptDir)

	//set up copier to overwrite pre-existing files, and sync them immediately
	opt := copy.Options{
		OnDirExists: func(src, dest string) copy.DirExistsAction {
			return 0
		},
		Sync: true,
	}
	copy.Copy(sDir, absScriptDir, opt)

}

type script struct {
	f *os.File
	*bufio.Scanner
}

func Open(sname string) (*script, error) {

	//add .sql if file given is not an sql file
	if filepath.Ext(sname) != ".sql" {
		sname = sname + ".sql"
	}

	sPath := filepath.Join(absScriptDir, sname)

	//open file
	f, err := os.OpenFile(sPath, os.O_RDONLY, os.ModePerm)
	if err != nil {

		//if file does not exist
		if os.IsNotExist(err) {

			//create file
			f, err = os.Create(sPath)
			if err != nil {
				return nil, err

			} else { //write todo Comment in file
				f.WriteString("-- TODO")
				f.Close()

				f, err = os.OpenFile(sPath, os.O_RDONLY, os.ModePerm)
				if err != nil {
					return nil, err
				}

			}

		} else {
			return nil, err
		}
	}

	sc := bufio.NewScanner(f)
	sc.Split(splitDelim(sqlDelim))

	s := &script{
		f:       f,
		Scanner: sc,
	}

	return s, nil
}

func splitDelim(delim string) (f func(data []byte, atEOF bool) (advance int, token []byte, err error)) {

	f = func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		var nCorr int = 0
		for i, b := range data {
			if b == delim[nCorr] {
				nCorr++

				if nCorr == len(delim) {
					return i + 1, data[:i-len(delim)], nil
				}
			}
		}

		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), data, nil
		}

		// request more data
		return 0, nil, nil
	}

	return
}

func (s *script) Next() bool {
	return s.Scan()
}

func (s *script) Stmt() string {
	return s.Text()
}

func (s *script) Close() error {
	return s.f.Close()
}

func (s *script) Exec(tx *sql.Tx) (sql.Result, error) {
	return tx.Exec(s.Stmt())
}

func (s *script) ExecAll(tx *sql.Tx) error {
	for s.Next() {
		_, err := s.Exec(tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func ExecScript(tx *sql.Tx, name string) error {
	s, err := Open(name)
	if err != nil {
		return err
	}
	defer s.Close()

	err = s.ExecAll(tx)
	if err != nil {
		return ScriptError{err}
	}

	return nil
}

//--- Schema opening

func UsingSchema(schema string) (*sql.Tx, error) {

	tx, err := DB.Begin()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(fmt.Sprintf("USE `%[1]s`", schema))
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func SchemaExists(schema string) (bool, error) {
	r, err := DB.Query("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", schema)
	if err != nil {
		return false, err
	}
	return r.Next(), nil
}

func MakeSchema(schema string) error {
	_, err := DB.Exec(fmt.Sprintf("CREATE DATABASE `%s` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;", schema))
	return err
}

func DelSchema(schema string) error {
	_, err := DB.Exec(fmt.Sprintf("DROP DATABASE `%s`;", schema))
	return err
}
