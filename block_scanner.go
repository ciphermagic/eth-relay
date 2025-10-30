package main

import (
	"encoding/json"
	"errors"
	"eth-relay/dao"
	"eth-relay/model"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

type BlockScanner struct {
	ethRequester ETHRPCRequester
	mysql        dao.MySQLConnector
	lastBlock    *dao.Block
	lastNumber   *big.Int
	fork         bool
	stop         chan bool
	lock         sync.Mutex
}

func NewBlockScanner(ethRequester ETHRPCRequester, mysql dao.MySQLConnector) *BlockScanner {
	return &BlockScanner{
		ethRequester: ethRequester,
		mysql:        mysql,
		lastBlock:    &dao.Block{},
		fork:         false,
		stop:         make(chan bool),
		lock:         sync.Mutex{},
	}
}

func (s *BlockScanner) Start() error {
	s.lock.Lock()
	if err := s.init(); err != nil {
		s.lock.Unlock()
		return err
	}
	execute := func() {
		if err := s.scan(); err != nil {
			s.log(err.Error())
			return
		}
		time.Sleep(1 * time.Second)
	}
	go func() {
		for {
			select {
			case <-s.stop:
				s.log("block scanner stopped")
				return
			default:
				if !s.fork {
					execute()
					continue
				}
				if err := s.init(); err != nil {
					s.log(err.Error())
					return
				}
				s.fork = false
			}
		}
	}()
	return nil
}

func (s *BlockScanner) init() error {
	_, err := s.mysql.Db.Desc("create_time").Where("fork=?", false).Get(s.lastBlock)
	if err != nil {
		return err
	}
	if s.lastBlock.BlockHash == "" {
		latestBlockNumber, err := s.ethRequester.GetLastestBlockNumber()
		if err != nil {
			return err
		}
		lastestBlock, err := s.ethRequester.GetBlockInfoByNumber(latestBlockNumber)
		if err != nil {
			return err
		}
		if lastestBlock.Number == "" {
			panic(latestBlockNumber.String())
		}
		s.lastBlock.BlockHash = lastestBlock.Hash
		s.lastBlock.ParentHash = lastestBlock.ParentHash
		s.lastBlock.BlockNumber = lastestBlock.Number
		s.lastBlock.CreateTime = s.hexToTen(lastestBlock.Timestamp).Int64()
		s.lastNumber = latestBlockNumber
	} else {
		s.lastNumber, _ = new(big.Int).SetString(s.lastBlock.BlockNumber, 10)
		s.lastNumber.Add(s.lastNumber, new(big.Int).SetInt64(1))
	}
	return nil
}

func (s *BlockScanner) getScannerBlockNumber() (*big.Int, error) {
	newBlockNumber, err := s.ethRequester.GetLastestBlockNumber()
	if err != nil {
		return nil, err
	}
	targetNumber := s.lastNumber
	if newBlockNumber.Cmp(s.lastNumber) < 0 {
	Next:
		for {
			select {
			case <-time.After(4 * time.Second):
				number, err := s.ethRequester.GetLastestBlockNumber()
				if err == nil && number.Cmp(s.lastNumber) >= 0 {
					targetNumber = number
					break Next
				}
			}
		}
	}
	return targetNumber, nil
}

func (s *BlockScanner) scan() error {
	targetNumber, err := s.getScannerBlockNumber()
	if err != nil {
		return err
	}
	fullBlock, err := s.retryGetBlockInfoByNumber(targetNumber)
	if err != nil {
		return err
	}
	s.lastNumber.Add(s.lastNumber, new(big.Int).SetInt64(1))

	tx := s.mysql.Db.NewSession()
	defer tx.Close()

	block := dao.Block{}
	_, err = tx.Where("block_hash=?", fullBlock.Hash).Get(&block)
	if err == nil && block.Id == 0 {
		block.BlockNumber = s.hexToTen(fullBlock.Number).String()
		block.ParentHash = fullBlock.ParentHash
		block.CreateTime = s.hexToTen(fullBlock.Timestamp).Int64()
		block.BlockHash = fullBlock.Hash
		block.Fork = false
		if _, err := tx.Insert(&block); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if s.forkCheck(&block) {
		data, _ := json.Marshal(fullBlock)
		s.log("分叉！", string(data))
		_ = tx.Commit()
		s.fork = true
		return errors.New("fork check")
	}
	s.log("scan block start ==> ", "number: ", s.hexToTen(fullBlock.Number), "hash: ", fullBlock.Hash)

	// 业务处理
	for index, transaction := range fullBlock.Transactions {
		s.log("tx hash ==> ", transaction.Hash)
		if index == 5 {
			break
		}
	}

	s.log("scan block finish\n===============================")

	if _, err = tx.Insert(&fullBlock.Transactions); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *BlockScanner) hexToTen(hex string) *big.Int {
	if !strings.HasPrefix(hex, "0x") {
		ten, _ := new(big.Int).SetString(hex, 10)
		return ten
	}
	ten, _ := new(big.Int).SetString(hex[2:], 16)
	return ten
}

func (s *BlockScanner) forkCheck(currentBlock *dao.Block) bool {
	if currentBlock.BlockNumber == "" {
		panic("invalid block")
	}
	if s.lastBlock.BlockHash == currentBlock.BlockHash || s.lastBlock.BlockHash == currentBlock.ParentHash {
		s.lastBlock = currentBlock
		return false
	}
	// 获取出最初开始分叉的那个区块
	forkBlock, err := s.getStartForkBlock(currentBlock.ParentHash)
	if err != nil {
		panic(err)
	}
	s.lastBlock = forkBlock // 更新。从这个区块开始，其之后的都是分叉的

	// 修改数据库记录，将分叉区块标记好
	numberEnd := ""
	if strings.HasPrefix(currentBlock.BlockNumber, "0x") {
		c, _ := new(big.Int).SetString(currentBlock.BlockNumber[2:], 16)
		numberEnd = c.String()
	} else {
		c, _ := new(big.Int).SetString(currentBlock.BlockNumber, 10)
		numberEnd = c.String()
	}
	numberFrom := forkBlock.BlockNumber
	_, err = s.mysql.Db.
		Table(dao.Block{}).
		Where("block_number > ? and block_number <= ?", numberFrom, numberEnd). // 区块号范围内
		Update(map[string]bool{"fork": true})
	if err != nil {
		panic(fmt.Errorf("update fork block failed %s", err.Error()))
	}
	return true
}

func (s *BlockScanner) getStartForkBlock(parentHash string) (*dao.Block, error) {
	parent := dao.Block{}
	_, err := s.mysql.Db.Where("block_hash=?", parentHash).Get(&parent)
	if err == nil && parent.BlockNumber != "" {
		return &parent, nil
	}
	parentFull, err := s.retryGetBlockInfoByHash(parentHash)
	if err != nil {
		return nil, fmt.Errorf("分叉严重错误，需要重启区块扫描 %s", err.Error())
	}
	return s.getStartForkBlock(parentFull.ParentHash)
}

func (s *BlockScanner) retryGetBlockInfoByHash(hash string) (*model.FullBlock, error) {
Retry:
	fullBlock, err := s.ethRequester.GetBlockInfoByHash(hash)
	if err != nil {
		errInfo := err.Error()
		if strings.Contains(errInfo, "empty") {
			s.log("获取区块信息，重试一次......", hash)
			goto Retry
		}
		return nil, err
	}
	return fullBlock, nil
}

func (s *BlockScanner) retryGetBlockInfoByNumber(targetNumber *big.Int) (*model.FullBlock, error) {
Retry:
	fullBlock, err := s.ethRequester.GetBlockInfoByNumber(targetNumber)
	if err != nil {
		errInfo := err.Error()
		if strings.Contains(errInfo, "empty") {
			s.log("获取区块信息，重试一次......", targetNumber.String())
			goto Retry
		}
		return nil, err
	}
	return fullBlock, nil
}

func (s *BlockScanner) log(args ...interface{}) {
	fmt.Println(args...)
}
