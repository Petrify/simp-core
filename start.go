package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"database/sql"

	term "github.com/Petrify/go-terminal"
	"github.com/Petrify/simpbot/config"

	_ "github.com/go-sql-driver/mysql"
)

const cfgPath string = "server_config.yml"

func init() {
	initFlags()
}

func main() {

	var (
		sConfig *cfg = newCfg()
	)

	fmt.Println(sConfig.DBType)

	flag.Parse()
	config.LoadCfg(sConfig, cfgPath)
	fmt.Println(sConfig.DBType)
	t := term.SysTerminal
	db, err := sql.Open(sConfig.DBType, sConfig.DBLogin)

	t.Println("Hello, Simp")
	t.Println(db.Stats())

	if err != nil {
		panic(err)
	}

	fmt.Println("Simp System Online. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	// Wait here until CTRL-C or other term signal is received.
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	t.Print(config.SaveCfg(sConfig, cfgPath))
}

type cfg struct {
	DBLogin string
	DBType  string
}

func newCfg() (c *cfg) {
	c = &cfg{}
	c.DBType = "mysql"

	return
}
