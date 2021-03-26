package simpsql

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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
	scanner *delimScanner
	last    string
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

	s := &script{
		scanner: newScanner(sqlDelim, f),
	}

	return s, nil
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

func (s *script) Stmt() string {
	return s.last
}

func (s *script) Close() error {
	return s.scanner.close()
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

	return ScriptError{s.ExecAll(tx)}
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
	_, err := DB.Exec("CREATE DATABASE :schema /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;",
		sql.Named("schema", schema))
	return err
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
