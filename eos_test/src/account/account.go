package account

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"sync"

	"bfc/src/account/wallet"
	"bfc/src/db_api"
	"bfc/src/utils"
)

type Wallet *wallet.Wallet


var AccountInDbMux sync.Mutex

type Account struct {
	Address      []byte
	MyWallet     Wallet
	Balance      float64
	Votes        uint64
	Pledge       uint64
	IsSubsidized bool
}

type AccountInDb struct {
	PublicKey    []byte
	Balance      float64
	Votes        uint64
	Pledge       uint64
	IsSubsidized bool
}

var GAccount *Account

// For Test , delete when test is over
func RegNewAccount() (*Account, error) {
	var newAccount Account
	myWallet := wallet.RegNewWallet()
	myAddress := myWallet.GetAddress()

	newAccount.Balance = 0
	newAccount.Votes = 0
	newAccount.Pledge = 0
	newAccount.Address = myAddress
	newAccount.MyWallet = myWallet
	newAccount.IsSubsidized = false
	// Save info to leveldb
	saveOk, err := SaveAccountInfoToDb(newAccount)
	if saveOk == true {
		return &newAccount, nil
	} else {
		return nil, err
	}
}

/*
* [Function]: Create an account
* [ input  ]: nil
* [ Return ]: *Account
 */
func NewAccount() (*Account, error) {
	var newAccount Account
	myWallet := wallet.NewWallet()
	myAddress := myWallet.GetAddress()
	_ = ioutil.WriteFile("./address.txt", myAddress, 0666)

	newAccount.Balance = 0
	newAccount.Votes = 0
	newAccount.Pledge = 0
	newAccount.Address = myAddress
	newAccount.MyWallet = myWallet
	newAccount.IsSubsidized = false
	// Save info to leveldb
	saveOk, err := SaveAccountInfoToDb(newAccount)
	if saveOk == true {
		return &newAccount, nil
	} else {
		return nil, err
	}
}

/*
* [Function]: Read the account info from local file, when use this func,you need
*			  to make that there are pem file in local path.
* [ input  ]: nil
* [ Return ]: *Account
 */
func InitAccountFromLocalFile() (*Account, error) {
	var localAccount Account
	// Read the pub/pri key from the .pem file
	privateKey := GetEccPrivateKey(wallet.PRIVATE_PATH)
	myWallet := wallet.InitWalletFromLocalFile(*privateKey)
	myAddress := myWallet.GetAddress()
	// reduction the account info from the level db
	// TODO : What if the user delete the db file or the file has been broken ?
	accountInDb, err := RecoverAccountInfoFromDb(string(myAddress))
	if accountInDb != nil && err == nil {
		localAccount.Balance = accountInDb.Balance
		localAccount.Votes = accountInDb.Votes
		localAccount.Pledge = accountInDb.Pledge
		localAccount.IsSubsidized = accountInDb.IsSubsidized
		localAccount.Address = myAddress
		localAccount.MyWallet = myWallet
		SaveAccountInfoToDb(localAccount)
		return &localAccount, nil
	} else {
		localAccount.Balance = 0
		localAccount.Votes = 0
		localAccount.Pledge = 0
		localAccount.Address = myAddress
		localAccount.MyWallet = myWallet
		localAccount.IsSubsidized = false
		return &localAccount, err
	}
}

/*
* [Function]: Read the balance info from leveldb
* [ input  ]: account's address (string)
* [ Return ]: *AccountInDb
 */
func RecoverAccountInfoFromDb(address string) (*AccountInDb, error) {
	if db_api.DbBlock == nil {
		return nil, nil
	}
	accountInDb := AccountInDb{}
	jsonAccount, _ := db_api.Db_getAccount(address)
	err := json.Unmarshal(jsonAccount, &accountInDb)
	if err == nil {
		return &accountInDb, nil
	} else {
		return nil, err
	}
}

/*
* [Function]: Save the account info from Account struct
* [ input  ]: account.Account
* [ Return ]: bool
 */
func SaveAccountInfoToDb(myAccount Account) (bool, error) {
	var accountDb AccountInDb
	accountDb.PublicKey = myAccount.MyWallet.PublicKey
	accountDb.Balance = myAccount.Balance
	accountDb.Votes = myAccount.Votes
	accountDb.Pledge = myAccount.Pledge
	accountDb.IsSubsidized = myAccount.IsSubsidized
	err2 := db_api.Db_updateAllAddress(string(myAccount.Address))
	jsonAccount, err := json.Marshal(accountDb)
	if err != nil {
		utils.Log.Errorf("AccountInfo serilized failed: %s", err.Error())
	}
	err = db_api.Db_insertAccount(jsonAccount, string(myAccount.Address))
	if err == nil && err2 == nil {
		return true, nil
	} else {
		if err2 != nil {
			utils.Log.Errorf("update address failed in SaveAccountInfoToDb, error:", err2.Error())
		}
		return false, err
	}
}

/*
* [Function]: Check if the publicKey.pem file is existance
* [ input  ]: nil
* [ Return ]: bool true->exist/false->not exist , error
 */
func CheckPublicPemExistance() bool {
	publicPem := wallet.PUBLIC_PATH

	_, err := os.Stat(publicPem)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}

	return false
}

/*
* [Function]: Check if the privateKey.pem file is existance
* [ input  ]: nil
* [ Return ]: bool true->exist/false->not exist , error
 */
func CheckPrivatePemExistance() bool {
	privatePem := wallet.PRIVATE_PATH

	_, err := os.Stat(privatePem)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}

	return false
}

/*
* [Function]: Use private key to sign
* [ input  ]: dataHash, the privateKey.pem file path
* [ Return ]: r + s signature
 */
func GetEccSignFromHashData(dataHash []byte, privateKeyPath string) []byte {
	// Get the private key from wallet
	privateKey := &GAccount.MyWallet.PrivateKey

	// Use private key sign the hash, return the sign(two big.Int)
	r, s, _ := ecdsa.Sign(rand.Reader, privateKey, dataHash)
	rText, _ := r.MarshalText()
	sText, _ := s.MarshalText()

	// Unit the rText and sText
	var sign bytes.Buffer
	sign.Write(rText)
	sign.Write([]byte(`+`))
	sign.Write(sText)

	return sign.Bytes()
}

/*
* [Function]: Use private key to sign
* [ input  ]: string plaintext, the privateKey.pem file path
* [ Return ]: r + s signature
 */
func GetEccSignFromPlainTextString(plainText string, privateKeyPath string) []byte {

	//turn the string to hash to sign
	dataHash, _ := utils.GetHashFromBytes([]byte(plainText))
	// Get the private key from wallet
	privateKey := &GAccount.MyWallet.PrivateKey

	// Use private key sign the hash, return the sign(two big.Int)
	r, s, _ := ecdsa.Sign(rand.Reader, privateKey, dataHash)
	rText, _ := r.MarshalText()
	sText, _ := s.MarshalText()

	// Unit the rText and sText
	var sign bytes.Buffer
	sign.Write(rText)
	sign.Write([]byte(`+`))
	sign.Write(sText)

	return sign.Bytes()
}

/*
* [Function]: Use the public key to check the sign.
* [ input  ]: bPublicKey ([]byte class from the wallet.marshal(*ecdsa.PublicKey)), dataHash, rText, sText
* [ Return ]: bool true/false
 */
func EccVerifyFromHashedData(bPublicKey, dataHash, sign []byte) bool {
	var result bool
	// Check the contract ID, if ID is not right return error
	hasAdd := strings.Index(string(sign), "+")
	if hasAdd == -1 {
		return false
	}
	// Get the public key
	publicKey := wallet.UnmarshalPubkey(bPublicKey)

	var r, s big.Int
	rs := bytes.Split(sign, []byte("+"))
	r.UnmarshalText(rs[0])
	s.UnmarshalText(rs[1])

	// Verify the signature
	if publicKey != nil {
		result = ecdsa.Verify(publicKey, dataHash, &r, &s)
	} else {
		result = false
	}
	return result
}

/*
* [Function]: Get the private key from the privateKey.pem
* [ input  ]: privateKey.pem file path
* [ Return ]: *ecdsa.PrivateKey in privateKey.pem
 */
func GetEccPrivateKey(path string) *ecdsa.PrivateKey {
	// Read the data from .pem
	file, _ := os.Open(path)
	fileInfo, _ := file.Stat()
	buf := make([]byte, fileInfo.Size())
	_, err := file.Read(buf)
	if err != nil {
		panic("fatal error , read the local private key failed , the process going to exit")
	}
	defer file.Close()

	// pem decode
	block, _ := pem.Decode(buf)

	// x509 decode
	privateKey, _ := x509.ParseECPrivateKey(block.Bytes)

	return privateKey
}

/*
* [Function]: Get the public key from the publicKey.pem
* [ input  ]: publicKey.pem file path
* [ Return ]: *ecdsa.PublicKey in privateKey.pem
 */
func GetEccPublicKey(path string) *ecdsa.PublicKey {
	// Read the data from publicKey.pem file path
	file, _ := os.Open(path)
	fileInfo, _ := file.Stat()
	buf := make([]byte, fileInfo.Size())
	file.Read(buf)
	defer file.Close()

	// pem解码
	block, _ := pem.Decode(buf)

	// x509解码
	publicKeyInterface, _ := x509.ParsePKIXPublicKey(block.Bytes)

	return publicKeyInterface.(*ecdsa.PublicKey)
}

func TransferToAccountInDb(account Account) AccountInDb {
	var accountInDb AccountInDb
	accountInDb.Balance = account.Balance
	accountInDb.Pledge = account.Pledge
	accountInDb.PublicKey = account.MyWallet.PublicKey
	accountInDb.IsSubsidized = account.IsSubsidized
	accountInDb.Votes = account.Votes
	return accountInDb
}

func GetAccountAsStructureFromDB(Address string) (AccountInDb, error) {
	jsonAccount, err := db_api.Db_getAccount(Address)
	if err != nil {
		utils.Log.Errorf("get account json in getting slice of address failed: %s", err.Error())
		return AccountInDb{}, err
	} else {
		var accountStruct AccountInDb
		err = json.Unmarshal(jsonAccount, &accountStruct)
		if err != nil {
			utils.Log.Errorf("account json unmarshal failed: %s", err.Error())
			return AccountInDb{}, err
		} else {
			return accountStruct, nil
		}
	}
}
