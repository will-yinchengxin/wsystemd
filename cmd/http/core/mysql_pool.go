package core

import (
	"errors"
	"fmt"
	"github.com/go-kit/kit/log/level"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	newLog "log"
	"os"
	"time"
	"wsystemd/cmd/log"
)

type MysqlConn struct {
	Host            string        `yaml:"host,omitempty"`
	Port            int           `yaml:"port,omitempty"`
	User            string        `yaml:"user,omitempty"`
	PassWord        string        `yaml:"password,omitempty"`
	DataBase        string        `yaml:"dataBase,omitempty"`
	MaxIdleConns    int           `yaml:"maxIdleConns,omitempty"`
	MaxOpenConns    int           `yaml:"maxOpenConns,omitempty"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime,omitempty"`
	ConnMaxIdletime time.Duration `yaml:"connMaxIdletime,omitempty"`
}

var (
	mysqlConfig map[string]interface{}
	mysqlPool   map[string]*gorm.DB
	DB_VRW      = "wsystemd"
)

func InitMysql() func() {
	mysqlConfig, err := GetMapConfig(CoreConfig, "mysql", MysqlConn{})
	if err != nil {
		panic(err)
	}
	if len(mysqlConfig) < 1 {
		panic("init gorm pool config failed, mysql config not found")
	}

	mysqlPool = make(map[string]*gorm.DB)
	for name, val := range mysqlConfig {
		if db, err := initGorm(val.(*MysqlConn)); err == nil && db != nil {
			mysqlPool[name] = db
		}
	}

	level.Info(log.Logger).Log("msg", "Init Mysql Pool Success!")

	return func() {
		for key, value := range mysqlPool {
			sqlDb, err := value.DB()
			err = sqlDb.Close()
			if err != nil {
				level.Error(log.Logger).Log("msg", "MySQL[ "+key+" ]Closed Err:"+err.Error())
				continue
			}
			level.Info(log.Logger).Log("msg", "MySQL[ "+key+" ]Closed Success!")
		}

	}
}

func initGorm(conn *MysqlConn) (*gorm.DB, error) {
	connectStr := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		conn.User, conn.PassWord, conn.Host, conn.Port, conn.DataBase)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       connectStr,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: logger.New(
			newLog.New(os.Stdout, "\r\n", newLog.LstdFlags),
			logger.Config{
				SlowThreshold: time.Second * 3,
				LogLevel:      logger.Silent,
				Colorful:      false,
			},
		),
	})
	if err != nil {
		logrus.Fatal(err.Error())
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(conn.MaxIdleConns)
	sqlDB.SetMaxOpenConns(conn.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Second * conn.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(time.Second * conn.ConnMaxIdletime)
	return db, nil
}

func GetDB(key string) (db *gorm.DB, err error) {
	db, ok := mysqlPool[key]
	if !ok {
		if config, ok := mysqlConfig[key]; !ok {
			return db, errors.New(key + " dbConfig doesn't exist")
		} else {
			if db, err = initGorm(config.(*MysqlConn)); err != nil {
				return db, errors.New(key + " dbConfig Initialization failure")
			} else {
				mysqlPool[key] = db
			}
		}
	}
	return
}
