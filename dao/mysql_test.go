package dao

import (
	"fmt"
	"testing"
)

func TestNewMqSQLConnector(t *testing.T) {
	option := MysqlOptions{
		DSN:                "root:123@tcp(127.0.0.1:3306)/eth_relay?charset=utf8mb4",
		TablePrefix:        "eth_",
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    15,
		ShowSqlLog:         true,
	}
	var tables []interface{}
	tables = append(tables, Block{}, Transaction{})
	mysql := NewMqSQLConnector(&option, tables)
	if mysql.Db.Ping() == nil {
		fmt.Println("数据库连接成功")
	} else {
		fmt.Println("数据库连接失败")
	}
}
