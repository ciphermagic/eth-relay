package main

import (
	"errors"
	"eth-relay/model"
	"eth-relay/tool"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
)

type ETHRPCRequester struct {
	nonceManager *NonceManager
	client       *ETHRPCClient
}

type ERC20BalanceRpcReq struct {
	ContractAddress string // 合约的以太坊地址
	UserAddress     string // 用户的以太坊地址
	ContractDecimal int    // 合约所对应代币的数位
}

func NewETHRPCRequester(nodeUrl string) *ETHRPCRequester {
	requester := &ETHRPCRequester{}
	requester.client = NewETHRPCClient(nodeUrl)
	requester.nonceManager = NewNonceManager()
	return requester
}

func NewETHWalletRequester() *ETHRPCRequester {
	requester := &ETHRPCRequester{}
	return requester
}

func (r *ETHRPCRequester) GetTransactionByHash(txHash string) (model.Transaction, error) {
	name := "eth_getTransactionByHash"
	res := model.Transaction{}
	err := r.client.GetRpc().Call(&res, name, txHash)
	return res, err
}

func (r *ETHRPCRequester) GetTransactions(txHashArr []string) ([]*model.Transaction, error) {
	name := "eth_getTransactionByHash"
	var resArr []*model.Transaction
	var reqs []rpc.BatchElem
	for _, txHash := range txHashArr {
		res := model.Transaction{}
		req := rpc.BatchElem{
			Method: name,
			Args:   []interface{}{txHash},
			Result: &res,
		}
		reqs = append(reqs, req)
		resArr = append(resArr, &res)
	}
	err := r.client.GetRpc().BatchCall(reqs)
	return resArr, err
}

func (r *ETHRPCRequester) GetETHBalance(address string) (string, error) {
	name := "eth_getBalance"
	res := ""
	err := r.client.GetRpc().Call(&res, name, address, "latest")
	if err != nil {
		return "", err
	}
	ten, _ := new(big.Int).SetString(res[2:], 16)
	return ten.String(), nil
}

func (r *ETHRPCRequester) GetEthBalances(addressArr []string) ([]string, error) {
	name := "eth_getBalance"
	var resArr []*string
	var reqs []rpc.BatchElem
	for _, addr := range addressArr {
		res := ""
		req := rpc.BatchElem{
			Method: name,
			Args:   []interface{}{addr, "latest"},
			Result: &res,
		}
		reqs = append(reqs, req)
		resArr = append(resArr, &res)
	}
	err := r.client.GetRpc().BatchCall(reqs)
	if err != nil {
		return nil, err
	}
	for _, req := range reqs {
		if req.Error != nil {
			return nil, req.Error
		}
	}
	var finalRes []string
	for _, item := range resArr {
		ten, _ := new(big.Int).SetString((*item)[2:], 16)
		finalRes = append(finalRes, ten.String())
	}
	return finalRes, err
}

func (r *ETHRPCRequester) GetERC20Balances(paramArr []ERC20BalanceRpcReq) ([]string, error) {
	name := "eth_call"
	methodId := "0x70a08231"
	var resArr []*string
	var reqs []rpc.BatchElem
	for _, param := range paramArr {
		res := ""
		arg := &model.CallArg{}
		userAddress := param.UserAddress
		arg.Gas = hexutil.EncodeUint64(30000)
		arg.To = common.HexToAddress(param.ContractAddress)
		arg.Data = methodId + "000000000000000000000000" + userAddress[2:]
		req := rpc.BatchElem{
			Method: name,
			Args:   []interface{}{arg, "latest"},
			Result: &res,
		}
		reqs = append(reqs, req)
		resArr = append(resArr, &res)
	}
	err := r.client.GetRpc().BatchCall(reqs)
	if err != nil {
		return nil, err
	}
	for _, req := range reqs {
		if req.Error != nil {
			return nil, req.Error
		}
	}
	var finalRes []string
	for _, item := range resArr {
		if *item == "" {
			continue
		}
		ten, _ := new(big.Int).SetString((*item)[2:], 16)
		finalRes = append(finalRes, ten.String())
	}
	return finalRes, err
}

func (r *ETHRPCRequester) GetLastestBlockNumber() (*big.Int, error) {
	name := "eth_blockNumber"
	number := ""
	err := r.client.GetRpc().Call(&number, name)
	if err != nil {
		return nil, err
	}
	ten, _ := new(big.Int).SetString(number[2:], 16)
	return ten, nil
}

func (r *ETHRPCRequester) GetBlockInfoByNumber(blockNumber *big.Int) (*model.FullBlock, error) {
	number := fmt.Sprintf("%#x", blockNumber)
	name := "eth_getBlockByNumber"
	fullBlock := &model.FullBlock{}
	err := r.client.GetRpc().Call(fullBlock, name, number, true)
	if err != nil {
		return nil, err
	}
	if fullBlock.Number == "" {
		return nil, errors.New("block info is empty")
	}
	return fullBlock, nil
}

func (r *ETHRPCRequester) GetBlockInfoByHash(blockHash string) (*model.FullBlock, error) {
	name := "eth_getBlockByHash"
	fullBlock := &model.FullBlock{}
	err := r.client.GetRpc().Call(fullBlock, name, blockHash, true)
	if err != nil {
		return nil, err
	}
	if fullBlock.Number == "" {
		return nil, errors.New("block info is empty")
	}
	return fullBlock, nil
}

func (r *ETHRPCRequester) ETHCall(request interface{}, arg model.CallArg) error {
	name := "eth_call"
	err := r.client.GetRpc().Call(request, name, arg, "latest")
	if err != nil {
		return err
	}
	return nil
}

func (r *ETHRPCRequester) CreateETHWallet(password string) (string, error) {
	if password == "" {
		return "", errors.New("password is empty")
	}
	if len(password) < 6 {
		return "", errors.New("password is too short")
	}
	keysDir := "./keystores"
	ks := keystore.NewKeyStore(keysDir, keystore.StandardScryptN, keystore.StandardScryptP)
	wallet, err := ks.NewAccount(password)
	if err != nil {
		return "0x", err
	}
	return wallet.Address.String(), nil
}

func (r *ETHRPCRequester) SendTransaction(address string, transaction *types.Transaction) (string, error) {
	signTx, err := tool.SignETHTransaction(address, transaction)
	if err != nil {
		return "", err
	}
	txRlpData, err := rlp.EncodeToBytes(&signTx)
	if err != nil {
		return "", err
	}
	txHash := ""
	name := "eth_sendRawTransaction"
	err = r.client.GetRpc().Call(&txHash, name, hexutil.Encode(txRlpData))
	if err != nil {
		return "", err
	}
	oldNonce := r.nonceManager.nonceMemCache[address]
	if oldNonce == nil {
		r.nonceManager.SetNonce(address, new(big.Int).SetUint64(transaction.Nonce()))
	}
	r.nonceManager.PlusNonce(address)
	return txHash, nil
}

func (r *ETHRPCRequester) GetNonce(address string) (uint64, error) {
	name := "eth_getTransactionCount"
	nonce := ""
	err := r.client.GetRpc().Call(&nonce, name, address, "pending")
	if err != nil {
		return 0, err
	}
	n, _ := new(big.Int).SetString(nonce[2:], 16)
	return n.Uint64(), nil
}

func (r *ETHRPCRequester) SendETHTransaction(fromStr, toStr, value string, gasLimit, gasPrice uint64) (string, error) {
	_to := common.HexToAddress(toStr)
	_gasPrice := new(big.Int).SetUint64(gasPrice)
	_value := tool.GetRealDecimalValue(value, 18)
	_amount, _ := new(big.Int).SetString(_value, 10)

	nonce := r.nonceManager.GetNonce(fromStr)
	if nonce == nil {
		n, err := r.GetNonce(fromStr)
		if err != nil {
			return "", err
		}
		nonce = new(big.Int).SetUint64(n)
		r.nonceManager.SetNonce(fromStr, nonce)
	}

	transaction := types.NewTx(&types.LegacyTx{
		Nonce:    nonce.Uint64(),
		GasPrice: _gasPrice,
		Gas:      gasLimit,
		To:       &_to,
		Value:    _amount,
		Data:     []byte(""),
	})
	return r.SendTransaction(fromStr, transaction)
}

func (r *ETHRPCRequester) SendERC20Transaction(fromStr, contract, receiver, valueStr string,
	gasLimit, gasPrice uint64, decimal int) (string, error) {
	_to := common.HexToAddress(contract)
	_gasPrice := new(big.Int).SetUint64(gasPrice)
	_amount := new(big.Int).SetInt64(0)

	nonce := r.nonceManager.GetNonce(fromStr)
	if nonce == nil {
		n, err := r.GetNonce(fromStr)
		if err != nil {
			return "", err
		}
		nonce = new(big.Int).SetUint64(n)
		r.nonceManager.SetNonce(fromStr, nonce)
	}

	data := tool.BuildERC20TransferData(valueStr, receiver, decimal)
	dataBytes := common.FromHex(data)

	transaction := types.NewTx(&types.LegacyTx{
		Nonce:    nonce.Uint64(),
		GasPrice: _gasPrice,
		Gas:      gasLimit,
		To:       &_to,
		Value:    _amount,
		Data:     dataBytes,
	})
	return r.SendTransaction(fromStr, transaction)
}
