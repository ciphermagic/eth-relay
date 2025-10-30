package main

import (
	"flag"
	"fmt"
	"os"

	"eth-relay/dao"
)

func main() {
	// ---------- 1. 命令行参数 ----------
	rpcURL := flag.String("rpc", "", "Ethereum JSON-RPC endpoint (http/https)")
	mysqlDSN := flag.String("mysql", "", "MySQL DSN, e.g. root:123@tcp(127.0.0.1:3306)/eth_relay?charset=utf8mb4")
	flag.Parse()

	// ---------- 2. 环境变量兜底 ----------
	if *rpcURL == "" {
		*rpcURL = os.Getenv("ETH_RPC_URL")
	}
	if *mysqlDSN == "" {
		*mysqlDSN = os.Getenv("MYSQL_DSN")
	}

	// ---------- 3. 参数校验 ----------
	if *rpcURL == "" {
		fmt.Println("Error: --rpc or ETH_RPC_URL must be provided")
		flag.Usage()
		os.Exit(1)
	}
	if *mysqlDSN == "" {
		fmt.Println("Error: --mysql or MYSQL_DSN must be provided")
		flag.Usage()
		os.Exit(1)
	}

	// ---------- 4. 初始化组件 ----------
	// MySQL
	mysqlOpt := dao.MysqlOptions{
		DSN:                *mysqlDSN,
		MaxOpenConnections: 20,
		MaxIdleConnections: 10,
		ConnMaxLifetime:    0,
		ShowSqlLog:         false,
		TablePrefix:        "eth_",
	}
	tables := []interface{}{dao.Block{}, dao.Transaction{}}
	mysqlConn := dao.NewMqSQLConnector(&mysqlOpt, tables)

	// ETH RPC
	requester := NewETHRPCRequester(*rpcURL)

	// Scanner
	scanner := NewBlockScanner(*requester, mysqlConn)

	// ---------- 5. 启动 ----------
	if err := scanner.Start(); err != nil {
		fmt.Println("Scanner start failed:", err)
		os.Exit(1)
	}

	// 阻塞主 goroutine
	select {}
}
