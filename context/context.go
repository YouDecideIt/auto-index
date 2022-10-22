package context

import (
	"database/sql"
	"github.com/YouDecideIt/auto-index/config"
)

type Context struct {
	Cfg *config.AutoIndexConfig
	DB  *sql.DB
}
