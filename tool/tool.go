package tool

import (
	"encoding/hex"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var ETHUnlockMap map[string]accounts.Account
var UnlockKs *keystore.KeyStore

func MakeMethodId(methodName string, abiStr string) (string, error) {
	_abi := &abi.ABI{}
	err := _abi.UnmarshalJSON([]byte(abiStr))
	if err != nil {
		return "", err
	}
	method := _abi.Methods[methodName]
	methodIdBytes := method.ID
	methodId := "0x" + common.Bytes2Hex(methodIdBytes)
	return methodId, nil
}

func UnlockETHWallet(keysDir, address, password string) error {
	if UnlockKs == nil {
		UnlockKs = keystore.NewKeyStore(keysDir, keystore.StandardScryptN, keystore.StandardScryptP)
		if UnlockKs == nil {
			return errors.New("unlock fail")
		}
	}
	unlock := accounts.Account{Address: common.HexToAddress(address)}
	if err := UnlockKs.Unlock(unlock, password); err != nil {
		return err
	}
	if ETHUnlockMap == nil {
		ETHUnlockMap = make(map[string]accounts.Account)
	}
	ETHUnlockMap[address] = unlock
	return nil
}

func ExportKeystore(keysDir, privateKeyHex, password string) error {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return err
	}
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return err
	}
	ks := keystore.NewKeyStore(keysDir, keystore.StandardScryptN, keystore.StandardScryptP)
	_, err = ks.ImportECDSA(privateKey, password)
	if err != nil {
		return err
	}
	return nil
}

func SignETHTransaction(address string, transaction *types.Transaction) (*types.Transaction, error) {
	if UnlockKs == nil {
		return nil, errors.New("you need to init keystore first")
	}
	account := ETHUnlockMap[address]
	if !common.IsHexAddress(account.Address.String()) {
		return nil, errors.New("account need to unlock first")
	}
	return UnlockKs.SignTx(account, transaction, nil) // 调用签名函数
}

func GetRealDecimalValue(value string, decimal int) string {
	if strings.Contains(value, ".") {
		// 小数
		arr := strings.Split(value, ".")
		if len(arr) != 2 {
			return ""
		}
		num := len(arr[1])
		left := decimal - num
		return arr[0] + arr[1] + strings.Repeat("0", left)
	} else {
		// 整数
		return value + strings.Repeat("0", decimal)
	}
}

func BuildERC20TransferData(value, receiver string, decimal int) string {
	realValue := GetRealDecimalValue(value, decimal)
	valueBig, _ := new(big.Int).SetString(realValue, 10)
	methodId := "0xa9059cbb"
	param1 := common.HexToHash(receiver).String()[2:]
	param2 := common.BytesToHash(valueBig.Bytes()).String()[2:]
	return methodId + param1 + param2
}
