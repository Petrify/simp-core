package discord

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// Bot Services
var (
	GCManager *GuildManager
	TManager  *TerminalManager
	DB        *sql.DB
)

func Run(token string) {

	TManager, _ = NewTerminalManager(token)
	GCManager, _ = NewGuildManager(token)
	DB, err := sql.Open("mysql", "root:LooT67@/practice")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}

}

func Close() {
	DB.Close()
}
