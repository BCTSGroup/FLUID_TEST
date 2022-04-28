package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"

	"bfc/src/utils"
	"bfc/src/utils/ripemd160"
)

const VERSION = byte(0x00)     //16进制0 版本号
const ADDRESS_CHECKSUM_LEN = 4 //地址检查长度4
const PRIVATE_PATH = "./privateKey.pem"
const PUBLIC_PATH = "./publicKey.pem"

/* 钱包结构 */
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

/*
* [Function]: 创建新钱包
* [ input  ]: nil
* [ Return ]: wallet 钱包结构,包含秘钥对
 */
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}
	wallet.saveKeyPair(PRIVATE_PATH, PUBLIC_PATH)
	return &wallet
}

// For Test, delete when test is over
func RegNewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}

/*
* [Function]: 从本地还原钱包
* [ input  ]: nil
* [ Return ]: wallet 钱包结构,包含秘钥对
 */
func InitWalletFromLocalFile(privateKey ecdsa.PrivateKey) *Wallet {
	public := MarshalPubkey(&privateKey.PublicKey)
	wallet := Wallet{privateKey, public}
	return &wallet
}

/*
* [Function]: 得到钱包地址
* [ input  ]: nil
* [ Return ]: address []byte 账户地址
 */
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)
	// 将版本号+pubKeyHash得到一个散列
	versionedPayload := append([]byte{VERSION}, pubKeyHash...)
	// 校验前4个字节的散列
	checksum := checksum(versionedPayload)
	// 将checksum和versionedPayload结合
	fullPayload := append(versionedPayload, checksum...)
	// BASE58 得到一个钱包地址
	address := utils.Base58Encode(fullPayload)
	return address
}

/*
* [Function]: Save the key pair to pem file
* [ input  ]: pri path, pub path
* [ Return ]: (bool) true->success false->wrong
 */
func (w Wallet) saveKeyPair(privatePath, publicPath string) {
	// Save the private key
	privateKey := w.PrivateKey
	x509PrivateKey, _ := x509.MarshalECPrivateKey(&privateKey)
	block := pem.Block{
		Type:  "ecc private key",
		Bytes: x509PrivateKey,
	}
	privateKFile, _ := os.Create(privatePath)
	pem.Encode(privateKFile, &block)
	defer privateKFile.Close()

	//Save the public key
	x509PublicKey, _ := x509.MarshalPKIXPublicKey(&w.PrivateKey.PublicKey)
	publicBlock := pem.Block{
		Type:  "ecc public key",
		Bytes: x509PublicKey,
	}

	publicFile, _ := os.Create(publicPath)
	defer publicFile.Close()
	pem.Encode(publicFile, &publicBlock)
}

/*
* [Function]: 椭圆算法生成并返回私钥公钥对
*
* 私钥结构如下：
*    type PrivateKey struct {
* 	    PublicKey    // 公钥信息
*	    D *big.Int   // 私钥，256位二进制随机数
*    }
 */
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	// 获取 P256 曲线
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err.Error())
	}
	pubKey := MarshalPubkey(&private.PublicKey)
	return *private, pubKey
}

/*
* [Function]: change pubKey []byte -> ecdsa.PublicKey
* [ input  ]: pub []byte
* [ Return ]: ecdsa.PublicKey
 */
func UnmarshalPubkey(pub []byte) *ecdsa.PublicKey {
	x, y := elliptic.Unmarshal(elliptic.P256(), pub)
	if x == nil {
		return nil
	}
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
}

/*
* [Function]: change ecdsa.PublicKey -> pubKey []byte
* [ input  ]: pub *ecdsa.PublicKey
* [ Return ]: pubKey []byte
 */
func MarshalPubkey(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(elliptic.P256(), pub.X, pub.Y)
}

/*
* [Function]: 使用RIPEMD160(SHA256(PubKey))哈希算法得到hashPubKey
* [ input  ]: pubKey []byte (由椭圆生成算法的得到的公钥)
* [ Return ]: publicRIPEMD160 []byte (hash后的公钥)
 */
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)
	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err.Error())
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	return publicRIPEMD160
}

/*
* [Function]: SHA256(SHA256(payload))算法返回前4个字节
* [ input  ]: payload []byte
* [ Return ]: SHA256(SHA256(PubKeyHash))
 */
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:ADDRESS_CHECKSUM_LEN]
}

/*
* [Function]: 校验地址
* [ input  ]: address string
* [ Return ]: bool
 */
func ValidateAddress(address string) bool {
	pubKeyHash := utils.Base58Decode([]byte(address))
	// 校验和
	actualChecksum := pubKeyHash[len(pubKeyHash)-ADDRESS_CHECKSUM_LEN:]
	// 版本号
	version := pubKeyHash[0]
	// 公钥哈希
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-ADDRESS_CHECKSUM_LEN]
	// 理论上正确的校验和
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))
	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// For test, transfer *ecdsa.PrivateKey to []byte after Base58
func MarshalPrivKeyBase58(privKey *ecdsa.PrivateKey) []byte {
	bPrivKey, _ := x509.MarshalECPrivateKey(privKey)
	bPrivKey = utils.Base58Encode(bPrivKey)
	return bPrivKey
}

// For test, transfer []byte after Base58 encode to *ecdsa.PrivateKey
func UnMarshalPrivKeyBase58(bPrivKeyEncoded []byte) *ecdsa.PrivateKey {
	bPrivKeyDecoded := utils.Base58Decode(bPrivKeyEncoded)
	privKey, _ := x509.ParseECPrivateKey(bPrivKeyDecoded)
	return privKey
}
