package main

import (
	"encoding/json"
	"eth-relay/model"
	"eth-relay/tool"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const localUrl = "http://localhost:8545"
const sepoliaUrl = "https://sepolia.infura.io/v3/123"

func TestETHRPCRequester_GetTransactionByHash(t *testing.T) {
	txHash := "0xdba68e394b13ba81e6645d7f4bfeec950a8f7a881777d19ce19d6bff4524362d"
	txInfo, err := NewETHRPCRequester(sepoliaUrl).GetTransactionByHash(txHash)
	if err != nil {
		panic(err)
	}
	res, _ := json.Marshal(txInfo)
	fmt.Println(string(res))
}

func TestETHRPCRequester_GetTransactions(t *testing.T) {
	txHash1 := "0xdba68e394b13ba81e6645d7f4bfeec950a8f7a881777d19ce19d6bff4524362d"
	txHash2 := "0xdba68e394b13ba81e6645d7f4bfeec950a8f7a881777d19ce19d6bff45243aaa"
	txHash3 := "0xfb308dbe4049fc290ad171c6bf51b7c1eacee7f5aedee13265011adfb71addc8"
	txHashArr := []string{txHash1, txHash2, txHash3}
	txInfos, err := NewETHRPCRequester(sepoliaUrl).GetTransactions(txHashArr)
	if err != nil {
		panic(err)
	}
	bytes, _ := json.Marshal(txInfos)
	fmt.Println(string(bytes))
}

func TestETHRPCRequester_GetETHBalance(t *testing.T) {
	address := "0xeE9A7E064DdddB8db82bB5cEf9E884409E7273fE"
	res, err := NewETHRPCRequester(sepoliaUrl).GetETHBalance(address)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func TestETHRPCRequester_GetLastestBlockNumber(t *testing.T) {
	number, err := NewETHRPCRequester(sepoliaUrl).GetLastestBlockNumber()
	if err != nil {
		panic(err)
	}
	fmt.Println(number.String())
}

func TestETHRPCRequester_GetBlockInfoByNumber(t *testing.T) {
	number, _ := NewETHRPCRequester(sepoliaUrl).GetLastestBlockNumber()
	fmt.Println(number.String())
	fullBlock, err := NewETHRPCRequester(sepoliaUrl).GetBlockInfoByNumber(number)
	if err != nil {
		panic(err)
	}
	info, _ := json.Marshal(fullBlock)
	fmt.Println(string(info))
}

func TestETHRPCRequester_GetBlockInfoByHash(t *testing.T) {
	hash := "0xf9ac4c0f3f5f1ffbe471b93feeafe69a5865d93fbffa8cfe72c2393479ccb259"
	fullBlock, err := NewETHRPCRequester(sepoliaUrl).GetBlockInfoByHash(hash)
	if err != nil {
		panic(err)
	}
	info, _ := json.Marshal(fullBlock)
	fmt.Println(string(info))
}

func TestETHRPCRequester_ETHCall(t *testing.T) {
	contractAbi :=
		`
[
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "spender",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "approve",
		"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"stateMutability": "nonpayable",
		"type": "constructor"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "spender",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "allowance",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "needed",
				"type": "uint256"
			}
		],
		"name": "ERC20InsufficientAllowance",
		"type": "error"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "sender",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "balance",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "needed",
				"type": "uint256"
			}
		],
		"name": "ERC20InsufficientBalance",
		"type": "error"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "approver",
				"type": "address"
			}
		],
		"name": "ERC20InvalidApprover",
		"type": "error"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "receiver",
				"type": "address"
			}
		],
		"name": "ERC20InvalidReceiver",
		"type": "error"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "sender",
				"type": "address"
			}
		],
		"name": "ERC20InvalidSender",
		"type": "error"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "spender",
				"type": "address"
			}
		],
		"name": "ERC20InvalidSpender",
		"type": "error"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "owner",
				"type": "address"
			},
			{
				"indexed": true,
				"internalType": "address",
				"name": "spender",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "Approval",
		"type": "event"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "to",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "transfer",
		"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "from",
				"type": "address"
			},
			{
				"indexed": true,
				"internalType": "address",
				"name": "to",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "Transfer",
		"type": "event"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "from",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "to",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "transferFrom",
		"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "owner",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "spender",
				"type": "address"
			}
		],
		"name": "allowance",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "account",
				"type": "address"
			}
		],
		"name": "balanceOf",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "decimals",
		"outputs": [
			{
				"internalType": "uint8",
				"name": "",
				"type": "uint8"
			}
		],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "name",
		"outputs": [
			{
				"internalType": "string",
				"name": "",
				"type": "string"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "symbol",
		"outputs": [
			{
				"internalType": "string",
				"name": "",
				"type": "string"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "totalSupply",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	}
]
		`
	methodName := "balanceOf"
	methodId, err := tool.MakeMethodId(methodName, contractAbi)
	if err != nil {
		panic(err)
	}
	contract := "0x5FbDB2315678afecb367f032d93F642f64180aa3"
	args := model.CallArg{
		To:   common.HexToAddress(contract),
		Data: methodId + "000000000000000000000000" + "eE9A7E064DdddB8db82bB5cEf9E884409E7273fE",
		Gas:  hexutil.EncodeUint64(30000),
	}
	result := ""
	err = NewETHRPCRequester(localUrl).ETHCall(&result, args)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
	ten, _ := new(big.Int).SetString(result[2:], 16)
	fmt.Println(ten)
}

func TestETHRPCRequester_CreateETHWallet(t *testing.T) {
	address, err := NewETHWalletRequester().CreateETHWallet("12345678")
	if err != nil {
		panic(err)
	}
	fmt.Println(address)
}

func TestETHRPCRequester_GetNonce(t *testing.T) {
	address := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	nonce, err := NewETHRPCRequester(localUrl).GetNonce(address)
	if err != nil {
		panic(err)
	}
	fmt.Println(nonce)
}

func TestETHRPCRequester_GetEthBalances(t *testing.T) {
	address1 := "0x6dB7Ee9774Be5a16685241fCeF5d6f968d9b0259"
	address2 := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	addressArr := []string{address1, address2}
	res, err := NewETHRPCRequester(localUrl).GetEthBalances(addressArr)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
	// [6976000000000000 9999977099090695592882]
}

func TestETHRPCRequester_SendETHTransaction(t *testing.T) {
	from := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	to := "0x6dB7Ee9774Be5a16685241fCeF5d6f968d9b0259"
	value := "1000"
	gasLimit := uint64(21000)
	gasPrice := uint64(36000000000)
	err := tool.UnlockETHWallet("./keystores", from, "12345678")
	if err != nil {
		panic(err)
	}
	txHash, err := NewETHRPCRequester(localUrl).SendETHTransaction(from, to, value, gasLimit, gasPrice)
	if err != nil {
		panic(err)
	}
	fmt.Println(txHash)
}

func TestETHRPCRequester_GetERC20Balance(t *testing.T) {
	contract := "0xc6e7DF5E7b4f2A278906862b61205850344D4e7d"
	var params []ERC20BalanceRpcReq

	item1 := ERC20BalanceRpcReq{}
	item1.ContractAddress = contract
	item1.UserAddress = "0xeE9A7E064DdddB8db82bB5cEf9E884409E7273fE"
	item1.ContractDecimal = 2
	params = append(params, item1)

	item2 := ERC20BalanceRpcReq{}
	item2.ContractAddress = contract
	item2.UserAddress = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	item2.ContractDecimal = 2
	params = append(params, item2)

	res, err := NewETHRPCRequester(localUrl).GetERC20Balances(params)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func TestETHRPCRequester_SendERC20Transaction(t *testing.T) {
	from := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	contract := "0xc6e7DF5E7b4f2A278906862b61205850344D4e7d"
	amount := "10"
	decimal := 2
	receiver := "0xeE9A7E064DdddB8db82bB5cEf9E884409E7273fE"
	gasLimit := uint64(500000)
	gasPrice := uint64(36000000000)
	err := tool.UnlockETHWallet("./keystores", from, "12345678")
	if err != nil {
		panic(err)
	}
	txHash, err := NewETHRPCRequester(localUrl).SendERC20Transaction(from, contract, receiver, amount, gasLimit, gasPrice, decimal)
	if err != nil {
		panic(err)
	}
	fmt.Println(txHash)
}
