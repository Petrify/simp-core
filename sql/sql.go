package simpsql

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/otiai10/copy"
)

var scriptDir = "sql_scripts"
var baseDir string

func init() {
	exedir, _ := os.Executable()
	baseDir = filepath.Dir(exedir)
}

type script struct {
	scanner *delimScanner
	last    string
}

func Open(sname string) (*script, error) {

	//add .sql if file given is not an sql file
	if filepath.Ext(sname) != "sql" {
		sname = sname + ".sql"
	}

	sDir := filepath.Join(baseDir, scriptDir)
	sPath := filepath.Join(sDir, sname)

	err := os.Mkdir(sDir, os.ModePerm)

	//if this directory is new, copy base sql_scripts folder from this module's sql_scripts
	if err == nil {
		cloneSqlScripts(sDir)
	}

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
				f.Sync()

				s := &script{
					scanner: newScanner(";", f),
				}

				return s, nil
			}

		}

		return nil, err

	}

	s := &script{
		scanner: newScanner(";\n", f),
	}

	return s, nil
}

//Clones Base sql_scripts folder from this module to target
func cloneSqlScripts(target string) {
	//get path to this source
	_, b, _, _ := runtime.Caller(0)
	modDir := filepath.Dir(filepath.Dir(b))
	sDir := filepath.Join(modDir, scriptDir)
	copy.Copy(sDir, target)

}

func (s *script) Next() bool {
	read, err := s.scanner.Scan()
	s.last = strings.TrimSpace(read)
	if err == io.EOF {
		if s.last == "" {
			return false
		}
	}
	return true
}

func (s *script) Prepare(tx *sql.Tx) (*sql.Stmt, error) {
	return tx.Prepare(s.last)
}

func (s *script) Raw() string {
	return s.last
}

func (s *script) Close() error {
	return s.scanner.close()
}

func (s *script) Exec(db *sql.DB, args ...interface{}) (sql.Result, error) {
	return db.Exec(s.Raw(), args...)
}

func (s *script) ExecTx(tx *sql.Tx, args ...interface{}) (sql.Result, error) {
	return tx.Exec(s.Raw(), args...)
}

func (s *script) Query(db *sql.DB, args ...interface{}) (*sql.Rows, error) {
	return db.Query(s.Raw(), args...)
}

func (s *script) QueryTx(tx *sql.Tx, args ...interface{}) (*sql.Rows, error) {
	return tx.Query(s.Raw(), args...)
}

func (s *script) ExecSchema(db *sql.DB, schema string, args ...interface{}) (sql.Result, error) {
	tx, err := WithSchema(db, schema)
	if err != nil {
		return nil, err
	}

	return s.ExecTx(tx, args...)
}

func (s *script) QuerySchema(db *sql.DB, schema string, args ...interface{}) (*sql.Rows, error) {
	tx, err := WithSchema(db, schema)
	if err != nil {
		return nil, err
	}

	return s.QueryTx(tx, args...)
}

func QueryScript(db *sql.DB, scriptName string, args ...interface{}) (*sql.Rows, error) {
	s, err := Open(scriptName)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	if !s.Next() {
		return nil, errors.New("empty sql file")
	}

	return db.Query(s.Raw(), args...)
}

func QueryScriptTx(tx *sql.Tx, scriptName string, args ...interface{}) (*sql.Rows, error) {
	s, err := Open(scriptName)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	if !s.Next() {
		return nil, errors.New("empty sql file")
	}

	return tx.Query(s.Raw(), args...)
}

func ExecScriptSchema(db *sql.DB, scriptName string, schema string, args ...interface{}) error {
	tx, err := WithSchema(db, schema)
	if err != nil {
		return err
	}

	return ExecScriptTx(tx, scriptName, args...)
}

func QueryScriptSchema(db *sql.DB, scriptName string, schema string, args ...interface{}) (*sql.Rows, error) {
	tx, err := WithSchema(db, schema)
	if err != nil {
		return nil, err
	}

	return QueryScriptTx(tx, scriptName, args...)
}

func (s *script) ExecAll(db *sql.DB, args ...interface{}) error {
	for s.Next() {
		_, err := s.Exec(db, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *script) ExecAllTx(tx *sql.Tx, args ...interface{}) error {
	for s.Next() {
		_, err := s.ExecTx(tx, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func ExecScript(db *sql.DB, name string, args ...interface{}) error {
	s, err := Open(name)
	if err != nil {
		return err
	}
	defer s.Close()

	return s.ExecAll(db)
}

func ExecScriptTx(tx *sql.Tx, name string, args ...interface{}) error {
	s, err := Open(name)
	if err != nil {
		return err
	}
	defer s.Close()

	return s.ExecAllTx(tx)
}

//--- Schema opening

func WithSchema(db *sql.DB, schema string) (*sql.Tx, error) {

	ok, err := SchemaExists(db, schema)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	if !ok {
		//schema does not exist
		_, err = tx.Exec(fmt.Sprintf("CREATE DATABASE `%[1]s` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;", schema))
		if err != nil {
			return nil, err
		}
	}

	_, err = tx.Exec(fmt.Sprintf("USE `%[1]s`", schema))
	if err != nil {
		return nil, err
	}

	return tx, nil

}

func SchemaExists(db *sql.DB, schema string) (bool, error) {
	r, err := db.Query("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", schema)
	if err != nil {
		return false, err
	}
	return r.Next(), nil
}

//--- Files

type delimScanner struct {
	f       *os.File
	delim   []byte
	bufSize int
	prevBuf [][]byte
}

func newScanner(delim string, f *os.File) *delimScanner {
	return &delimScanner{
		f:       f,
		delim:   []byte(delim),
		bufSize: 2048,
		prevBuf: make([][]byte, 0, 8),
	}
}

func (s *delimScanner) Scan() (string, error) {

	// if there are no previous buffers
	if len(s.prevBuf) == 0 {
		err := s.readNext()
		if err != nil {
			return "", err
		}
	}

	nCorr := 0 //number of correct hits
	for {
		lastbuf := s.prevBuf[len(s.prevBuf)-1] //latest buffer
		for i := range lastbuf {
			if lastbuf[i] == s.delim[nCorr] {
				nCorr++
				if nCorr == len(s.delim) {
					//Delim found
					return s.compileString(i-len(s.delim), true), nil
				}
			} else {
				nCorr = 0
			}
		}

		if len(lastbuf) == s.bufSize {
			err := s.readNext()
			if err != nil {
				if err == io.EOF {
					return s.compileString(s.bufSize-1, false), err
				}
				return "", err
			}
		} else {
			return s.compileString(len(lastbuf)-1, false), io.EOF
		}
	}
}

func (s *delimScanner) readNext() error {
	newBuf := make([]byte, s.bufSize)
	n, err := s.f.Read(newBuf)
	if err != nil {
		return err
	}
	s.prevBuf = append(s.prevBuf, newBuf[:n])
	return nil
}

func (s *delimScanner) compileString(lbIndex int, rmDelim bool) string {
	nbuf := len(s.prevBuf)
	lTotal := (nbuf-1)*s.bufSize + lbIndex + 1
	remain := lTotal
	bytes := make([]byte, 0, lTotal)
	i := 0
	for ; remain > s.bufSize; i++ {
		bytes = append(bytes, s.prevBuf[i]...)
		remain -= s.bufSize
	}
	bytes = append(bytes, s.prevBuf[i][:remain]...)

	//reset scanner buffer
	s.prevBuf = s.prevBuf[nbuf-1:]

	if rmDelim {
		s.prevBuf[0] = s.prevBuf[0][lbIndex+len(s.delim)+1:]
	} else {
		s.prevBuf[0] = s.prevBuf[0][lbIndex+1:]
	}

	return string(bytes)
}

func (s *delimScanner) close() error {
	return s.f.Close()
}
