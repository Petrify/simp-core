package simp

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"database/sql"

	term "github.com/Petrify/go-terminal"
	"github.com/Petrify/simp-core/config"
	"github.com/Petrify/simp-core/service"
	simpsql "github.com/Petrify/simp-core/sql"

	_ "github.com/go-sql-driver/mysql"
)

const cfgPathRelative string = "server_config.yml"

var (
	cfgPath string
	sConfig *cfg
	t       *term.Terminal
)

func Name() string {
	return sConfig.Name
}

func init() {
	start()
}

func start() {

	exeFile, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exeDir := filepath.Dir(exeFile)
	cfgPath = filepath.Join(exeDir, cfgPathRelative)

	sConfig = newCfg()
	config.LoadCfg(sConfig, cfgPath)

	// Terminal
	t = term.SysTerminal

	db, err := sql.Open("mysql", sConfig.DBLogin)
	if err != nil {
		t.Print("Database connection failed. Stopping.")
		os.Exit(0)
	}

	simpsql.DB = db

	service.StartSysService(t.Logger, sConfig.Name)

	if err != nil {
		panic(err)
	}

	t.Println("Simp System Online")
}

func Wait() {

	t.Println("Press Ctrl-C to exit")
	sc := make(chan os.Signal, 1)
	// Wait here until CTRL-C or other term signal is received.
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	config.SaveCfg(sConfig, cfgPath)
}

type cfg struct {
	DBLogin string
	Name    string
}

func newCfg() (c *cfg) {
	c = &cfg{}
	c.Name = "simp"
	c.DBLogin = "user:password@/dbname"

	return
}
