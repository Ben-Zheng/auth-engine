package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/auth-engine/pkg/constants"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"golang.org/x/xerrors"
)

func GenerateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", xerrors.Errorf("%w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CheckTimeFormat checks if the given time string matches the "08:00:00" format.
func CheckTimeFormat(timeStr string) (bool, time.Time) {
	const format = constants.DaliyTimeFormat
	t, err := time.ParseInLocation(format, timeStr, time.Local)
	if err != nil {
		hlog.Errorf("CheckTimeFormat error: %v", err)
	}
	return err == nil, t
}

// CheckDateFormat checks if the given date string matches the "2006-01-02" format.
func CheckDateFormat(dateStr string) (bool, time.Time) {
	const format = constants.DateRangeTimeFormat
	d, err := time.ParseInLocation(format, dateStr, time.Local)
	if err != nil {
		hlog.Errorf("CheckDateFormat error: %v", err)
	}
	return err == nil, d
}

// CheckDateTimeFormat checks if the given datetime string matches the "2006-01-02 15:04:05" format.
func CheckDateTimeFormat(datetimeStr string) (bool, time.Time) {
	const format = constants.TimeFormat
	d, err := time.ParseInLocation(format, datetimeStr, time.Local)
	if err != nil {
		hlog.Errorf("CheckDateTimeFormat error: %v", err)
	}
	return err == nil, d
}

// encryptData 使用AES加密数据
func EncryptData(data []byte) ([]byte, error) {
	// Create a new AES block cipher
	block, err := aes.NewCipher([]byte(constants.AesCryptionKey))
	if err != nil {
		return nil, err
	}
	// Create a new GCM cipher mode instance
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	// Generate a new nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	// Encrypt the data
	encryptedData := gcm.Seal(nonce, nonce, data, nil)
	return encryptedData, nil
}

// decryptData 使用AES解密数据
func DecryptData(encryptedData []byte) ([]byte, error) {
	// Create a new AES block cipher
	block, err := aes.NewCipher([]byte(constants.AesCryptionKey))
	if err != nil {
		return nil, err
	}
	// Create a new GCM cipher mode instance
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
