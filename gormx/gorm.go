package gormx

import (
	"log/slog"
	"strings"

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

func New(driverName DriverType, dsn string, logger *slog.Logger) *DB {

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
		logger.Error("unknown gorm driver type", "driver", driverName)
		return nil
	}

	db, err := gorm.Open(dialector, &gorm.Config{})

	if err != nil {
		logger.Error("connect to db failed", "driver", driverName, "error", err)
		return nil
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
