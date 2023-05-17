package lib

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"errors"
)

type Encrypt interface {
	Type() byte
	Encode(data []byte, key []byte) ([]byte, error)
	Decode(data []byte, key []byte) ([]byte, error)
}

type aesEncrypt struct {
	mode AESMode
}

type AESMode string

const (
	CBC AESMode = "cbc"
	ECB AESMode = "ecb"
	CFB AESMode = "cfb"
)

var (
	ErrAESNotFound = errors.New("aes mode not found")
)

func NewAESEncrypt(mode AESMode) Encrypt {
	return &aesEncrypt{mode: mode}
}

func (a *aesEncrypt) Type() byte {
	return 0x00
}

func (a *aesEncrypt) Encode(data []byte, key []byte) ([]byte, error) {
	switch a.mode {
	case CBC:
		return a.cbcEncrypt(data, key)
	case CFB:
	case ECB:
	default:
		return nil, ErrAESNotFound
	}
	return nil, nil
}

func (a *aesEncrypt) Decode(data []byte, key []byte) ([]byte, error) {
	switch a.mode {
	case CBC:
		return a.cbcDecrypt(data, key)
	}
	return nil, nil
}

func (a *aesEncrypt) cbcEncrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	data = a.pkcs5Padding(data, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, []byte("1234567890ABCDEF"))
	encrypted := make([]byte, len(data))
	blockMode.CryptBlocks(encrypted, data)
	return encrypted, nil
}

func (a *aesEncrypt) cbcDecrypt(encrypted []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, []byte("1234567890ABCDEF"))
	decrypted := make([]byte, len(encrypted))
	blockMode.CryptBlocks(decrypted, encrypted)
	decrypted = a.pkcs5UnPadding(decrypted)
	return decrypted, nil
}

func (a *aesEncrypt) ecbEncrypt(data []byte, key []byte) ([]byte, error) {
	cipher, err := aes.NewCipher(a.generateKey(key))
	if err != nil {
		return nil, err
	}
	length := (len(data) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)
	copy(plain, data)
	pad := byte(len(plain) - len(data))
	for i := len(data); i < len(plain); i++ {
		plain[i] = pad
	}
	encrypted := make([]byte, len(plain))
	for bs, be := 0, cipher.BlockSize(); bs <= len(data); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Encrypt(encrypted[bs:be], plain[bs:be])
	}
	return encrypted, nil
}

func (a *aesEncrypt) generateKey(key []byte) (genKey []byte) {
	genKey = make([]byte, 16)
	copy(genKey, key)
	for i := 16; i < len(key); {
		for j := 0; j < 16 && i < len(key); j, i = j+1, i+1 {
			genKey[j] ^= key[i]
		}
	}
	return genKey
}
func (a *aesEncrypt) pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func (a *aesEncrypt) pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

type rsaEncrypt struct {
}

func (r *rsaEncrypt) Encode(data []byte, key []byte) ([]byte, error) {
	panic("implement me")
}

func (r *rsaEncrypt) Decode(data []byte, key []byte) ([]byte, error) {
	panic("implement me")
}

// @brief:填充明文
func PKCS5Padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padtext...)
}

// @brief:去除填充数据
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// @brief:AES加密
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, []byte("1234567890ABCDEF")) //初始向量的长度必须等于块block的长度16字节
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// @brief:AES解密
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) //初始向量的长度必须等于块block的长度16字节
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

// 3des加密
type triple struct {
}

func TripAES() Encrypt {
	return &triple{}
}

func (a *triple) Type() byte {
	return 0x00
}

func (t *triple) Encode(data []byte, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	data = PKCS5Padding(data, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:8])
	crypt := make([]byte, len(data))
	blockMode.CryptBlocks(crypt, data)
	return crypt, nil
}

func (t *triple) Decode(data []byte, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	ctx := make([]byte, len(data))
	blockMode := cipher.NewCBCEncrypter(block, key[:8])
	blockMode.CryptBlocks(ctx, data)
	ctx = PKCS5UnPadding(ctx)
	return ctx, nil
}

type cbcEncrypt struct {
}

func NewCBCEncrypt() Encrypt {
	return &cbcEncrypt{}
}

func (a *cbcEncrypt) Type() byte {
	return 0x02
}

func (a *cbcEncrypt) Encode(data []byte, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	data = PKCS5Padding(data, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key)
	encrypted := make([]byte, len(data))
	blockMode.CryptBlocks(encrypted, data)
	return encrypted, nil
}

func (a *cbcEncrypt) Decode(encrypted []byte, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, key)
	decrypted := make([]byte, len(encrypted))
	blockMode.CryptBlocks(decrypted, encrypted)
	decrypted = PKCS5UnPadding(decrypted)
	return decrypted, nil
}

type ecbEncypt struct {
}

func NewECBEncrypt() Encrypt {
	return &ecbEncypt{}
}

func (a *ecbEncypt) Type() byte {
	return 0x03
}

func (a *ecbEncypt) Encode(data []byte, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bs := block.BlockSize()
	data = PKCS5Padding(data, bs)
	if len(data)%bs != 0 {
		return nil, errors.New("Need a multiple of the blocksize")
	}
	ciphertext := make([]byte, len(data))
	dst := ciphertext
	for len(data) > 0 {
		block.Encrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	return ciphertext, nil
}

func (a *ecbEncypt) Decode(data []byte, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	if len(data)%bs != 0 {
		return nil, errors.New("input not full blocks")
	}
	plaintext := make([]byte, len(data))
	dst := plaintext
	for len(data) > 0 {
		block.Decrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	plaintext = PKCS5UnPadding(plaintext)
	return plaintext, nil
}
