package hook

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
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dop251/goja"
)

// RegisterBuiltins 注册内置模块到 JS 运行时
func RegisterBuiltins(vm *goja.Runtime) {
	// crypto 模块 - 加密解密
	vm.Set("crypto", map[string]interface{}{
		"md5":              cryptoMD5,              // crypto.md5(data) - MD5 哈希
		"sha1":             cryptoSHA1,             // crypto.sha1(data) - SHA1 哈希
		"sha256":           cryptoSHA256,           // crypto.sha256(data) - SHA256 哈希
		"sha512":           cryptoSHA512,           // crypto.sha512(data) - SHA512 哈希
		"hmacMD5":          cryptoHmacMD5,          // crypto.hmacMD5(data, key) - HMAC-MD5
		"hmacSHA1":         cryptoHmacSHA1,         // crypto.hmacSHA1(data, key) - HMAC-SHA1
		"hmacSHA256":       cryptoHmacSHA256,       // crypto.hmacSHA256(data, key) - HMAC-SHA256
		"aesEncrypt":       cryptoAESEncrypt,       // crypto.aesEncrypt(plaintext, key) - AES-CBC加密(自动生成IV)
		"aesDecrypt":       cryptoAESDecrypt,       // crypto.aesDecrypt(ciphertext, key) - AES-CBC解密(自动提取IV)
		"aesCBCEncrypt":    cryptoAESCBCEncrypt,    // crypto.aesCBCEncrypt(plaintext, key, iv) - AES-CBC加密(自定义IV)
		"aesCBCDecrypt":    cryptoAESCBCDecrypt,    // crypto.aesCBCDecrypt(ciphertext, key, iv) - AES-CBC解密(自定义IV)
		"aesECBEncrypt":    cryptoAESECBEncrypt,    // crypto.aesECBEncrypt(plaintext, key) - AES-ECB加密
		"aesECBDecrypt":    cryptoAESECBDecrypt,    // crypto.aesECBDecrypt(ciphertext, key) - AES-ECB解密
		"desEncrypt":       cryptoDESEncrypt,       // crypto.desEncrypt(plaintext, key, iv) - DES-CBC加密(8字节key)
		"desDecrypt":       cryptoDESDecrypt,       // crypto.desDecrypt(ciphertext, key, iv) - DES-CBC解密
		"des3Encrypt":      cryptoDES3Encrypt,      // crypto.des3Encrypt(plaintext, key, iv) - 3DES-CBC加密(24字节key)
		"des3Decrypt":      cryptoDES3Decrypt,      // crypto.des3Decrypt(ciphertext, key, iv) - 3DES-CBC解密
		"rsaEncrypt":       cryptoRSAEncrypt,       // crypto.rsaEncrypt(plaintext, publicKeyPEM) - RSA公钥加密
		"rsaDecrypt":       cryptoRSADecrypt,       // crypto.rsaDecrypt(ciphertext, privateKeyPEM) - RSA私钥解密
		"rsaSign":          cryptoRSASign,          // crypto.rsaSign(data, privateKeyPEM) - RSA私钥签名(SHA256)
		"rsaVerify":        cryptoRSAVerify,        // crypto.rsaVerify(data, signature, publicKeyPEM) - RSA公钥验签
		"randomBytes":      cryptoRandomBytes,      // crypto.randomBytes(n) - 生成n字节随机数(返回hex)
	})

	// http 模块 - HTTP 请求
	vm.Set("http", map[string]interface{}{
		"get":      httpGet,      // http.get(url, headers) - GET请求
		"post":     httpPost,     // http.post(url, body, headers) - POST请求(原始body)
		"postJSON": httpPostJSON, // http.postJSON(url, data, headers) - POST JSON请求
		"postForm": httpPostForm, // http.postForm(url, data, headers) - POST 表单请求
		"request":  httpRequest,  // http.request(method, url, body, headers) - 自定义请求
	})

	// encoding 模块 - 编解码
	vm.Set("encoding", map[string]interface{}{
		"base64Encode": base64Encode, // encoding.base64Encode(data) - Base64编码
		"base64Decode": base64Decode, // encoding.base64Decode(data) - Base64解码
		"hexEncode":    hexEncode,    // encoding.hexEncode(data) - Hex编码
		"hexDecode":    hexDecode,    // encoding.hexDecode(data) - Hex解码
		"jsonEncode":   jsonEncode,   // encoding.jsonEncode(obj) - JSON序列化
		"jsonDecode":   jsonDecode,   // encoding.jsonDecode(str) - JSON反序列化
		"xmlEncode":    xmlEncode,    // encoding.xmlEncode(obj) - XML序列化
		"xmlDecode":    xmlDecode,    // encoding.xmlDecode(str) - XML反序列化
		"urlEncode":    urlEncode,    // encoding.urlEncode(str) - URL编码
		"urlDecode":    urlDecode,    // encoding.urlDecode(str) - URL解码
	})

	// util 模块 - 工具函数
	vm.Set("util", map[string]interface{}{
		"uuid":       utilUUID,       // util.uuid() - 生成UUID
		"now":        utilNow,        // util.now() - 当前时间戳(秒)
		"formatTime": utilFormatTime, // util.formatTime(timestamp, "YYYY-MM-DD HH:mm:ss") - 格式化时间
		"parseTime":  utilParseTime,  // util.parseTime(str, "YYYY-MM-DD HH:mm:ss") - 解析时间
		"sleep":      utilSleep,      // util.sleep(ms) - 休眠毫秒
	})

	// console 模块 - 控制台输出
	vm.Set("console", map[string]interface{}{
		"log": consoleLog, // console.log(...args) - 打印日志
	})
}

// ============ crypto 模块 ============

// cryptoMD5 计算 MD5 哈希
// 参数: data - 待哈希字符串
// 返回: 32位小写hex字符串
func cryptoMD5(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// cryptoSHA1 计算 SHA1 哈希
// 参数: data - 待哈希字符串
// 返回: 40位小写hex字符串
func cryptoSHA1(data string) string {
	hash := sha1.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// cryptoSHA256 计算 SHA256 哈希
// 参数: data - 待哈希字符串
// 返回: 64位小写hex字符串
func cryptoSHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// cryptoSHA512 计算 SHA512 哈希
// 参数: data - 待哈希字符串
// 返回: 128位小写hex字符串
func cryptoSHA512(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

// cryptoHmacMD5 计算 HMAC-MD5
// 参数: data - 待签名数据, key - 密钥
// 返回: 32位小写hex字符串
func cryptoHmacMD5(data, key string) string {
	h := hmac.New(md5.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// cryptoHmacSHA1 计算 HMAC-SHA1
// 参数: data - 待签名数据, key - 密钥
// 返回: 40位小写hex字符串
func cryptoHmacSHA1(data, key string) string {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// cryptoHmacSHA256 计算 HMAC-SHA256
// 参数: data - 待签名数据, key - 密钥
// 返回: 64位小写hex字符串
func cryptoHmacSHA256(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// cryptoAESEncrypt AES-CBC 加密(自动生成IV，IV拼接在密文前)
// 参数: plaintext - 明文, key - 16/24/32字节密钥
// 返回: Base64编码的密文(IV+密文)
func cryptoAESEncrypt(plaintext, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// PKCS7 padding
	blockSize := block.BlockSize()
	padding := blockSize - len(plaintext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	plainBytes := append([]byte(plaintext), padText...)

	// CBC 加密，IV放在密文前面
	ciphertext := make([]byte, blockSize+len(plainBytes))
	iv := ciphertext[:blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[blockSize:], plainBytes)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// cryptoAESDecrypt AES-CBC 解密(自动提取IV)
// 参数: ciphertext - Base64编码的密文(IV+密文), key - 16/24/32字节密钥
// 返回: 明文
func cryptoAESDecrypt(ciphertext, key string) (string, error) {
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

	// PKCS7 unpadding
	padding := int(data[len(data)-1])
	return string(data[:len(data)-padding]), nil
}

// cryptoAESCBCEncrypt AES-CBC 加密(自定义IV)
// 参数: plaintext - 明文, key - 16/24/32字节密钥, iv - 16字节IV
// 返回: Base64编码的密文
func cryptoAESCBCEncrypt(plaintext, key, iv string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// PKCS7 padding
	blockSize := block.BlockSize()
	padding := blockSize - len(plaintext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	plainBytes := append([]byte(plaintext), padText...)

	ciphertext := make([]byte, len(plainBytes))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, plainBytes)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// cryptoAESCBCDecrypt AES-CBC 解密(自定义IV)
// 参数: ciphertext - Base64编码的密文, key - 16/24/32字节密钥, iv - 16字节IV
// 返回: 明文
func cryptoAESCBCDecrypt(ciphertext, key, iv string) (string, error) {
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

	// PKCS7 unpadding
	padding := int(data[len(data)-1])
	return string(data[:len(data)-padding]), nil
}

// cryptoAESECBEncrypt AES-ECB 加密(不安全，仅兼容老系统)
// 参数: plaintext - 明文, key - 16/24/32字节密钥
// 返回: Base64编码的密文
func cryptoAESECBEncrypt(plaintext, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// PKCS7 padding
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

// cryptoAESECBDecrypt AES-ECB 解密
// 参数: ciphertext - Base64编码的密文, key - 16/24/32字节密钥
// 返回: 明文
func cryptoAESECBDecrypt(ciphertext, key string) (string, error) {
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

	// PKCS7 unpadding
	padding := int(plaintext[len(plaintext)-1])
	return string(plaintext[:len(plaintext)-padding]), nil
}

// cryptoDESEncrypt DES-CBC 加密
// 参数: plaintext - 明文, key - 8字节密钥, iv - 8字节IV
// 返回: Base64编码的密文
func cryptoDESEncrypt(plaintext, key, iv string) (string, error) {
	block, err := des.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// PKCS7 padding
	blockSize := block.BlockSize()
	padding := blockSize - len(plaintext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	plainBytes := append([]byte(plaintext), padText...)

	ciphertext := make([]byte, len(plainBytes))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, plainBytes)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// cryptoDESDecrypt DES-CBC 解密
// 参数: ciphertext - Base64编码的密文, key - 8字节密钥, iv - 8字节IV
// 返回: 明文
func cryptoDESDecrypt(ciphertext, key, iv string) (string, error) {
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

	// PKCS7 unpadding
	padding := int(data[len(data)-1])
	return string(data[:len(data)-padding]), nil
}

// cryptoDES3Encrypt 3DES-CBC 加密
// 参数: plaintext - 明文, key - 24字节密钥, iv - 8字节IV
// 返回: Base64编码的密文
func cryptoDES3Encrypt(plaintext, key, iv string) (string, error) {
	block, err := des.NewTripleDESCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// PKCS7 padding
	blockSize := block.BlockSize()
	padding := blockSize - len(plaintext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	plainBytes := append([]byte(plaintext), padText...)

	ciphertext := make([]byte, len(plainBytes))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, plainBytes)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// cryptoDES3Decrypt 3DES-CBC 解密
// 参数: ciphertext - Base64编码的密文, key - 24字节密钥, iv - 8字节IV
// 返回: 明文
func cryptoDES3Decrypt(ciphertext, key, iv string) (string, error) {
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

	// PKCS7 unpadding
	padding := int(data[len(data)-1])
	return string(data[:len(data)-padding]), nil
}

// cryptoRSAEncrypt RSA 公钥加密
// 参数: plaintext - 明文, publicKeyPEM - PEM格式公钥
// 返回: Base64编码的密文
func cryptoRSAEncrypt(plaintext, publicKeyPEM string) (string, error) {
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

// cryptoRSADecrypt RSA 私钥解密
// 参数: ciphertext - Base64编码的密文, privateKeyPEM - PEM格式私钥(支持PKCS1和PKCS8)
// 返回: 明文
func cryptoRSADecrypt(ciphertext, privateKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试 PKCS8
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

// cryptoRSASign RSA 私钥签名(SHA256)
// 参数: data - 待签名数据, privateKeyPEM - PEM格式私钥
// 返回: Base64编码的签名
func cryptoRSASign(data, privateKeyPEM string) (string, error) {
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

// cryptoRSAVerify RSA 公钥验签(SHA256)
// 参数: data - 原始数据, signature - Base64编码的签名, publicKeyPEM - PEM格式公钥
// 返回: 是否验证通过
func cryptoRSAVerify(data, signature, publicKeyPEM string) (bool, error) {
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

// cryptoRandomBytes 生成随机字节
// 参数: n - 字节数
// 返回: hex编码的随机字符串
func cryptoRandomBytes(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ============ http 模块 ============

// httpGet 发送 GET 请求
// 参数: urlStr - 请求URL, headers - 请求头(可选)
// 返回: {status: 状态码, headers: 响应头, body: 响应体, json: 解析后的JSON}
func httpGet(urlStr string, headers map[string]string) (map[string]interface{}, error) {
	return httpRequest("GET", urlStr, nil, headers)
}

// httpPost 发送 POST 请求(原始body)
// 参数: urlStr - 请求URL, body - 请求体, headers - 请求头(可选)
// 返回: {status: 状态码, headers: 响应头, body: 响应体, json: 解析后的JSON}
func httpPost(urlStr string, body string, headers map[string]string) (map[string]interface{}, error) {
	return httpRequest("POST", urlStr, body, headers)
}

// httpPostJSON 发送 POST JSON 请求
// 参数: urlStr - 请求URL, data - 对象(自动序列化为JSON), headers - 请求头(可选)
// 返回: {status: 状态码, headers: 响应头, body: 响应体, json: 解析后的JSON}
func httpPostJSON(urlStr string, data interface{}, headers map[string]string) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json"
	return httpRequest("POST", urlStr, string(jsonBytes), headers)
}

// httpPostForm 发送 POST 表单请求
// 参数: urlStr - 请求URL, data - 表单数据, headers - 请求头(可选)
// 返回: {status: 状态码, headers: 响应头, body: 响应体, json: 解析后的JSON}
func httpPostForm(urlStr string, data map[string]string, headers map[string]string) (map[string]interface{}, error) {
	form := url.Values{}
	for k, v := range data {
		form.Set(k, v)
	}
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	return httpRequest("POST", urlStr, form.Encode(), headers)
}

// httpRequest 发送自定义 HTTP 请求
// 参数: method - 请求方法, urlStr - 请求URL, body - 请求体, headers - 请求头
// 返回: {status: 状态码, headers: 响应头, body: 响应体, json: 解析后的JSON}
func httpRequest(method, urlStr string, body interface{}, headers map[string]string) (map[string]interface{}, error) {
	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case []byte:
			bodyReader = bytes.NewReader(v)
		default:
			jsonBytes, _ := json.Marshal(v)
			bodyReader = bytes.NewReader(jsonBytes)
		}
	}

	req, err := http.NewRequest(method, urlStr, bodyReader)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 尝试解析为 JSON
	var jsonData interface{}
	if err := json.Unmarshal(respBody, &jsonData); err == nil {
		return map[string]interface{}{
			"status":  resp.StatusCode,
			"headers": resp.Header,
			"body":    string(respBody),
			"json":    jsonData,
		}, nil
	}

	return map[string]interface{}{
		"status":  resp.StatusCode,
		"headers": resp.Header,
		"body":    string(respBody),
		"json":    nil,
	}, nil
}

// ============ encoding 模块 ============

// base64Encode Base64 编码
// 参数: data - 待编码字符串
// 返回: Base64编码后的字符串
func base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// base64Decode Base64 解码
// 参数: data - Base64编码的字符串
// 返回: 解码后的字符串
func base64Decode(data string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// hexEncode Hex 编码
// 参数: data - 待编码字符串
// 返回: Hex编码后的字符串
func hexEncode(data string) string {
	return hex.EncodeToString([]byte(data))
}

// hexDecode Hex 解码
// 参数: data - Hex编码的字符串
// 返回: 解码后的字符串
func hexDecode(data string) (string, error) {
	decoded, err := hex.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// jsonEncode JSON 序列化
// 参数: data - 待序列化对象
// 返回: JSON字符串
func jsonEncode(data interface{}) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// jsonDecode JSON 反序列化
// 参数: data - JSON字符串
// 返回: 解析后的对象
func jsonDecode(data string) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// xmlEncode XML 序列化
// 参数: data - 待序列化对象
// 返回: XML字符串
func xmlEncode(data interface{}) (string, error) {
	xmlBytes, err := xml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(xmlBytes), nil
}

// xmlDecode XML 反序列化(简单实现)
// 参数: data - XML字符串
// 返回: 扁平化的map，key为路径如"root.child"
func xmlDecode(data string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	decoder := xml.NewDecoder(strings.NewReader(data))
	var stack []string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := token.(type) {
		case xml.StartElement:
			stack = append(stack, t.Name.Local)
		case xml.CharData:
			if len(stack) > 0 {
				key := strings.Join(stack, ".")
				result[key] = strings.TrimSpace(string(t))
			}
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return result, nil
}

// urlEncode URL 编码
// 参数: data - 待编码字符串
// 返回: URL编码后的字符串
func urlEncode(data string) string {
	return url.QueryEscape(data)
}

// urlDecode URL 解码
// 参数: data - URL编码的字符串
// 返回: 解码后的字符串
func urlDecode(data string) (string, error) {
	return url.QueryUnescape(data)
}

// ============ util 模块 ============

// utilUUID 生成 UUID
// 返回: UUID字符串(格式: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
func utilUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// utilNow 获取当前时间戳
// 返回: Unix时间戳(秒)
func utilNow() int64 {
	return time.Now().Unix()
}

// utilFormatTime 格式化时间戳
// 参数: timestamp - Unix时间戳(秒), layout - 格式(支持YYYY/MM/DD/HH/mm/ss)
// 返回: 格式化后的时间字符串
func utilFormatTime(timestamp int64, layout string) string {
	layout = strings.ReplaceAll(layout, "YYYY", "2006")
	layout = strings.ReplaceAll(layout, "MM", "01")
	layout = strings.ReplaceAll(layout, "DD", "02")
	layout = strings.ReplaceAll(layout, "HH", "15")
	layout = strings.ReplaceAll(layout, "mm", "04")
	layout = strings.ReplaceAll(layout, "ss", "05")
	return time.Unix(timestamp, 0).Format(layout)
}

// utilParseTime 解析时间字符串
// 参数: timeStr - 时间字符串, layout - 格式(支持YYYY/MM/DD/HH/mm/ss)
// 返回: Unix时间戳(秒)
func utilParseTime(timeStr, layout string) (int64, error) {
	layout = strings.ReplaceAll(layout, "YYYY", "2006")
	layout = strings.ReplaceAll(layout, "MM", "01")
	layout = strings.ReplaceAll(layout, "DD", "02")
	layout = strings.ReplaceAll(layout, "HH", "15")
	layout = strings.ReplaceAll(layout, "mm", "04")
	layout = strings.ReplaceAll(layout, "ss", "05")
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// utilSleep 休眠
// 参数: ms - 毫秒数
func utilSleep(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// ============ console 模块 ============

// consoleLog 打印日志
// 参数: args - 任意参数
func consoleLog(args ...interface{}) {
	fmt.Println(args...)
}
