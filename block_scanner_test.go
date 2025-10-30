package main

import (
	"eth-relay/dao"
	"testing"
)

func TestBlockScanner_Start(t *testing.T) {
	option := dao.MysqlOptions{
		Hostname:           "localhost",
		Port:               "6034",
		DbName:             "eth_relay",
		User:               "root",
		Password:           "123",
		TablePrefix:        "eth_",
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    15,
		ShowSqlLog:         false,
	}
	var tables []interface{}
	tables = append(tables, dao.Block{}, dao.Transaction{})
	mysql := dao.NewMqSQLConnector(&option, tables)

	requester := NewETHRPCRequester(sepoliaUrl)
	scanner := NewBlockScanner(*requester, mysql)
	err := scanner.Start()
	if err != nil {
		panic(err)
	}
	select {}
}
