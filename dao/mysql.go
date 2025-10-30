package dao

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"xorm.io/core"
)

type MySQLConnector struct {
	options *MysqlOptions
	tables  []interface{}
	Db      *xorm.Engine // xorm 框架实例
}

type MysqlOptions struct {
	DSN                string
	TablePrefix        string // 数据库表前缀
	MaxOpenConnections int    // 数据库最大连接数
	MaxIdleConnections int    // 数据库最大空闲连接数
	ConnMaxLifetime    int    // 空闲链接空闲多久被回收，单位秒
	ShowSqlLog         bool
}

func NewMqSQLConnector(options *MysqlOptions, tables []interface{}) MySQLConnector {
	var connector MySQLConnector
	connector.options = options
	connector.tables = tables
	db, err := xorm.NewEngine("mysql", options.DSN)
	if err != nil {
		panic(fmt.Errorf("数据库初始化失败 %s", err.Error()))
	}
	tbMapper := core.NewPrefixMapper(core.SnakeMapper{}, options.TablePrefix)
	db.SetTableMapper(tbMapper)
	db.DB().SetConnMaxLifetime(time.Duration(options.ConnMaxLifetime) * time.Second)
	db.DB().SetMaxIdleConns(options.MaxIdleConnections)
	db.DB().SetMaxOpenConns(options.MaxOpenConnections)
	db.ShowSQL(options.ShowSqlLog)
	err = db.Ping()
	if err != nil {
		panic(fmt.Errorf("数据库连接失败 %s", err.Error()))
	}
	connector.Db = db

	if err := connector.createTables(); err != nil {
		panic(fmt.Errorf("创建数据表失败 %s", err.Error()))
	}
	return connector
}

func (s *MySQLConnector) createTables() error {
	if err := s.Db.CreateTables(s.tables...); err != nil {
		return fmt.Errorf("create mysql table error:%s", err.Error())
	}
	if err := s.Db.Sync2(s.tables...); err != nil {
		return fmt.Errorf("sync table error:%s", err.Error())
	}
	return nil
}
