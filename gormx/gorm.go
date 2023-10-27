package gormx

import (
	"strings"

	"github.com/c1emon/gcommon/logx"
	_ "github.com/lib/pq"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DriverType int

const (
	Unknown DriverType = iota - 1
	Postgres
	Mysql
	Sqlite
)

func (d DriverType) String() string {
	switch d {
	case Postgres:
		return "postgres"
	case Mysql:
		return "mysql"
	case Sqlite:
		return "sqlite"
	default:
		return "unknown"
	}
}

func ParseDriverType(dt string) DriverType {
	switch strings.ToLower(dt) {
	case "postgres":
		return Postgres
	case "mysql":
		return Mysql
	case "sqlite":
		return Sqlite
	default:
		return Unknown
	}
}

type DB struct {
	*gorm.DB
}

func New(driverName DriverType, dsn string, loggerFactory logx.LoggerFactory) *DB {
	logger := logx.NewGormLogger(loggerFactory)

	var dialector gorm.Dialector
	switch driverName {
	case Postgres:
		dialector = postgres.Open(dsn)
	case Mysql:
		dialector = mysql.Open(dsn)
	case Sqlite:
		dialector = sqlite.Open(dsn)
	case Unknown:
	default:
		logger.Panic("unknown driver type: %s", driverName)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger,
	})

	if err != nil {
		logger.Panic("unable connect to %s: %s", driverName, err)
	}

	return &DB{
		db,
	}
}

func (db *DB) Close() error {
	d, err := db.DB.DB()
	if err != nil {
		return err
	}
	return d.Close()
}
