package gormx

import (
	"strings"
	"sync"

	"github.com/c1emon/gcommon/logx"
	_ "github.com/lib/pq"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var once = sync.Once{}
var gormInstance *gorm.DB
var drv string
var src string

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

func Initialize(dbDrv, dbSrc string) {
	drv = dbDrv
	src = dbSrc
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

func connect(driverName DriverType, dsn string) {
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
		logx.GetLogger().Panicf("unknown driver type: %s", driverName)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logx.GetGormLogrusLogger(),
	})

	if err != nil {
		logx.GetLogger().Panicf("unable connect to %s: %s", driverName, err)
	}
	gormInstance = db
}

func GetGormDB() *gorm.DB {

	once.Do(func() {

		connect(ParseDriverType(drv), src)
	})

	return gormInstance
}

func DisConnect() error {
	if gormInstance != nil {
		d, err := gormInstance.DB()
		if err != nil {
			return nil
		}
		return d.Close()
	}
	return nil
}
