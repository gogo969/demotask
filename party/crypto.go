package party

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/sha1"
	"encoding/base64"

	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"

	"math/rand"
	"time"

	md5simd "github.com/minio/md5-simd"
)

// md5
func MD5Hash(text string) string {

	server := md5simd.NewServer()
	md5Hash := server.NewHash()
	_, _ = md5Hash.Write([]byte(text))
	digest := md5Hash.Sum([]byte{})
	encrypted := hex.EncodeToString(digest)

	server.Close()
	md5Hash.Close()

	return encrypted
}

// md5
func getMD5Hash(text string) string {

	server := md5simd.NewServer()
	md5Hash := server.NewHash()
	_, _ = md5Hash.Write([]byte(text))
	digest := md5Hash.Sum([]byte{})
	encrypted := hex.EncodeToString(digest)

	server.Close()
	md5Hash.Close()

	return encrypted
}

// sha1
func Sha1Sum(s string) []byte {

	h := sha1.New()
	h.Write([]byte(s))

	return h.Sum(nil)
}

// sha256
func sha256sum(param []byte) string {

	h := sha256.New()
	h.Write(param)

	return fmt.Sprintf("%x", h.Sum(nil))
}

func rsaEncrypt(privateKey, origData []byte) []byte {

	//设置私钥
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil
	}

	prkI, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil
	}

	priv := prkI.(*rsa.PrivateKey)
	encodeByte, _ := rsa.SignPKCS1v15(crand.Reader, priv, crypto.MD5, origData)

	return encodeByte
}

// aes ecd
func aesEcbEncrypt(data, key []byte) []byte {

	block, _ := aes.NewCipher(key)

	data = pkcs5Padding(data, block.BlockSize())
	decrypted := make([]byte, len(data))
	size := block.BlockSize()

	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		block.Encrypt(decrypted[bs:be], data[bs:be])
	}

	return decrypted
}

// aes ecd
func aesEcbDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	origData := make([]byte, len(crypted))

	size := block.BlockSize()

	for bs, be := 0, size; bs < len(crypted); bs, be = bs+size, be+size {
		block.Decrypt(origData[bs:be], crypted[bs:be])
	}

	origData = pkcs5UnPadding(origData)
	return origData, nil
}

func cryptBlocks(block cipher.Block, origData, crypted []byte) {

	for len(crypted) > 0 {
		block.Decrypt(origData, crypted[:block.BlockSize()])
		crypted = crypted[block.BlockSize():]
		origData = origData[block.BlockSize():]
	}
}

// aes cbc
func aesCbcEncrypt(plaintext []byte, secretKey, iv string) []byte {

	keyBytes := []byte(secretKey)
	aesBlockCipher, _ := aes.NewCipher(keyBytes)
	content := pkcs5Padding(plaintext, aesBlockCipher.BlockSize())
	encrypted := make([]byte, len(content))
	aesBlockMode := cipher.NewCBCEncrypter(aesBlockCipher, []byte(iv))
	aesBlockMode.CryptBlocks(encrypted, content)

	return encrypted
}

// des
func desEncrypt(src, key []byte) (string, error) {

	block, err := des.NewCipher(key)
	if err != nil {
		return "", err
	}

	bs := block.BlockSize()
	src = pkcs5Padding(src, bs)
	if len(src)%bs != 0 {
		return "", errors.New("need a multiple of the block size")
	}

	out := make([]byte, len(src))
	dst := out
	for len(src) > 0 {
		block.Encrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}

	return Base64Encode(out), nil
}

func desDecrypt(src, key []byte) ([]byte, error) {

	if len(src) < 1 {
		return nil, errors.New("src nil")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	out := make([]byte, len(src))
	dst := out
	bs := block.BlockSize()
	if len(src)%bs != 0 {
		return nil, errors.New("crypto/cipher: input not full blocks")
	}

	for len(src) > 0 {
		block.Decrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	out = pkcs5UnPadding(out)

	return out, nil
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {

	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(origData []byte) []byte {

	length := len(origData)
	unPadding := int(origData[length-1])

	if length-unPadding < 0 {
		return origData
	}

	return origData[:(length - unPadding)]
}

func randomString(length int) string {

	b := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	var rd []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		rd = append(rd, b[r.Intn(len(b))])
	}

	return string(rd)
}

// 3DES加密
func tripleDesEncrypt(data, key []byte) ([]byte, error) {

	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}

	data = pkcs5Padding(data, block.BlockSize())
	decrypted := make([]byte, len(data))
	size := block.BlockSize()

	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		block.Encrypt(decrypted[bs:be], data[bs:be])
	}

	return decrypted, nil
}

func Base64Encode(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

//获取source的子串,如果start小于0或者end大于source长度则返回""
//start:开始index，从0开始，包括0
//end:结束index，以end结束，但不包括end
func substring(source string, start int, end int) string {

	var r = []rune(source)
	length := len(r)

	if start < 0 || end > length || start > end {
		return ""
	}

	if start == 0 && end == length {
		return source
	}

	return string(r[start:end])
}
