package simp

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"database/sql"

	term "github.com/Petrify/go-terminal"
	"github.com/Petrify/simp-core/config"
	"github.com/Petrify/simp-core/service"

	_ "github.com/go-sql-driver/mysql"
)

const cfgPathRelative string = "server_config.yml"

var cfgPath string

func init() {
	initFlags()
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cfgPath = cwd + "/" + cfgPathRelative
}

func Start() {

	var (
		sConfig *cfg = newCfg()
	)

	fmt.Println(sConfig.DBType)

	flag.Parse()
	config.LoadCfg(sConfig, cfgPath)

	t := term.SysTerminal

	//Declare variables for DB connection loop
	const maxRetries = 5
	var (
		db  *sql.DB
		err error
	)

	//Connect to DB. Retry if necessary
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("mysql", sConfig.DBLogin)
		if err != nil {
			t.Printf("Error opening Database connection \n%e \nRetrying... [%d/%d]", err, i+1, maxRetries)
		}
	}

	if err != nil {
		t.Println("Database connection failed. Stopping.")
		panic(err)
	}

	service.StartSysService(db, t.Logger)

	if err != nil {
		panic(err)
	}

	t.Println("Simp System Online. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	// Wait here until CTRL-C or other term signal is received.
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	config.SaveCfg(sConfig, cfgPath)
}

type cfg struct {
	DBLogin string
	DBType  string
}

func newCfg() (c *cfg) {
	c = &cfg{}
	c.DBType = "mysql"
	c.DBLogin = "user:password@/dbname"

	return
}
