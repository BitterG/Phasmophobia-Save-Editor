package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// es3Decrypt 解密 EasySave3 存档数据
// 对应 Python: es3_decrypt(enc_data, key)
// 格式：前16字节为 IV/salt，其余为 AES-CBC 密文，PKCS7 padding
func es3Decrypt(encData []byte, key []byte) ([]byte, error) {
	if len(encData) < 32 {
		return nil, fmt.Errorf("加密数据太短")
	}

	// 取前16字节作为 salt（也用作 IV）
	salt := encData[:16]
	ciphertext := encData[16:]

	// PBKDF2-SHA1，迭代100次，生成16字节密钥
	derivedKey := pbkdf2.Key(key, salt, 100, 16, sha1.New)

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("创建 AES cipher 失败: %w", err)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("密文长度不是 AES 块大小的整数倍")
	}

	mode := cipher.NewCBCDecrypter(block, salt)
	decrypted := make([]byte, len(ciphertext))
	mode.CryptBlocks(decrypted, ciphertext)

	// 去除 PKCS7 padding
	padLen := int(decrypted[len(decrypted)-1])
	if padLen == 0 || padLen > aes.BlockSize {
		return nil, fmt.Errorf("无效的 PKCS7 padding: %d", padLen)
	}
	decrypted = decrypted[:len(decrypted)-padLen]

	return decrypted, nil
}

// es3Encrypt 加密数据为 EasySave3 存档格式
// 对应 Python: es3_encrypt(dec_data, key)
func es3Encrypt(decData []byte, key []byte) ([]byte, error) {
	// 随机生成16字节 IV（同时用作 PBKDF2 salt）
	iv := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("生成随机 IV 失败: %w", err)
	}

	derivedKey := pbkdf2.Key(key, iv, 100, 16, sha1.New)

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("创建 AES cipher 失败: %w", err)
	}

	// PKCS7 padding
	paddingLen := aes.BlockSize - (len(decData) % aes.BlockSize)
	padded := make([]byte, len(decData)+paddingLen)
	copy(padded, decData)
	for i := len(decData); i < len(padded); i++ {
		padded[i] = byte(paddingLen)
	}

	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	// 输出格式：IV(16字节) + 密文
	result := append(iv, ciphertext...)
	return result, nil
}
