package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
)

// MD5 计算 MD5 哈希
func MD5(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SHA1 计算 SHA1 哈希
func SHA1(data string) string {
	hash := sha1.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SHA256 计算 SHA256 哈希
func SHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SHA512 计算 SHA512 哈希
func SHA512(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

// HmacMD5 计算 HMAC-MD5
func HmacMD5(data, key string) string {
	h := hmac.New(md5.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// HmacSHA1 计算 HMAC-SHA1
func HmacSHA1(data, key string) string {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// HmacSHA256 计算 HMAC-SHA256
func HmacSHA256(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// AESEncrypt AES-CBC 加密(自动生成IV，IV拼接在密文前)
func AESEncrypt(plaintext, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	padding := blockSize - len(plaintext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	plainBytes := append([]byte(plaintext), padText...)

	ciphertext := make([]byte, blockSize+len(plainBytes))
	iv := ciphertext[:blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[blockSize:], plainBytes)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt AES-CBC 解密(自动提取IV)
func AESDecrypt(ciphertext, key string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	if len(data) < blockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := data[:blockSize]
	data = data[blockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)

	padding := int(data[len(data)-1])
	return string(data[:len(data)-padding]), nil
}

// AESCBCEncrypt AES-CBC 加密(自定义IV)
func AESCBCEncrypt(plaintext, key, iv string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	padding := blockSize - len(plaintext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	plainBytes := append([]byte(plaintext), padText...)

	ciphertext := make([]byte, len(plainBytes))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, plainBytes)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESCBCDecrypt AES-CBC 解密(自定义IV)
func AESCBCDecrypt(ciphertext, key, iv string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	mode.CryptBlocks(data, data)

	padding := int(data[len(data)-1])
	return string(data[:len(data)-padding]), nil
}

// AESECBEncrypt AES-ECB 加密
func AESECBEncrypt(plaintext, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	padding := blockSize - len(plaintext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	plainBytes := append([]byte(plaintext), padText...)

	ciphertext := make([]byte, len(plainBytes))
	for i := 0; i < len(plainBytes); i += blockSize {
		block.Encrypt(ciphertext[i:i+blockSize], plainBytes[i:i+blockSize])
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESECBDecrypt AES-ECB 解密
func AESECBDecrypt(ciphertext, key string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	plaintext := make([]byte, len(data))
	for i := 0; i < len(data); i += blockSize {
		block.Decrypt(plaintext[i:i+blockSize], data[i:i+blockSize])
	}

	padding := int(plaintext[len(plaintext)-1])
	return string(plaintext[:len(plaintext)-padding]), nil
}

// DESEncrypt DES-CBC 加密
func DESEncrypt(plaintext, key, iv string) (string, error) {
	block, err := des.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	padding := blockSize - len(plaintext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	plainBytes := append([]byte(plaintext), padText...)

	ciphertext := make([]byte, len(plainBytes))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, plainBytes)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DESDecrypt DES-CBC 解密
func DESDecrypt(ciphertext, key, iv string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := des.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	mode.CryptBlocks(data, data)

	padding := int(data[len(data)-1])
	return string(data[:len(data)-padding]), nil
}

// DES3Encrypt 3DES-CBC 加密
func DES3Encrypt(plaintext, key, iv string) (string, error) {
	block, err := des.NewTripleDESCipher([]byte(key))
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	padding := blockSize - len(plaintext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	plainBytes := append([]byte(plaintext), padText...)

	ciphertext := make([]byte, len(plainBytes))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, plainBytes)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DES3Decrypt 3DES-CBC 解密
func DES3Decrypt(ciphertext, key, iv string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := des.NewTripleDESCipher([]byte(key))
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	mode.CryptBlocks(data, data)

	padding := int(data[len(data)-1])
	return string(data[:len(data)-padding]), nil
}

// RSAEncrypt RSA 公钥加密
func RSAEncrypt(plaintext, publicKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("not RSA public key")
	}

	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, []byte(plaintext))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// RSADecrypt RSA 私钥解密
func RSADecrypt(ciphertext, privateKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", err
		}
		priv = privKey.(*rsa.PrivateKey)
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, priv, data)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// RSASign RSA 私钥签名(SHA256)
func RSASign(data, privateKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", err
		}
		priv = privKey.(*rsa.PrivateKey)
	}

	hash := sha256.Sum256([]byte(data))
	signature, err := rsa.SignPKCS1v15(rand.Reader, priv, 0, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// RSAVerify RSA 公钥验签(SHA256)
func RSAVerify(data, signature, publicKeyPEM string) (bool, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return false, fmt.Errorf("failed to parse PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("not RSA public key")
	}

	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}

	hash := sha256.Sum256([]byte(data))
	err = rsa.VerifyPKCS1v15(rsaPub, 0, hash[:], sig)
	return err == nil, nil
}

// RandomBytes 生成随机字节
func RandomBytes(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
