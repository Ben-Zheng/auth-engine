package test

import (
	"crypto/rand"
	"encoding/json"
	"testing"

	"github.com/auth-engine/internal/pkg/dao"
	"github.com/auth-engine/internal/pkg/models"
	"github.com/auth-engine/internal/pkg/services"
	"github.com/auth-engine/internal/pkg/utils"
)

var (
	db                      = dao.GetDB()
	iTokenDao               = dao.NewTokenDao(db)
	iTokenValidityPolicyDao = dao.NewTokenValidityPolicyDao(db)
	tokenService            = services.NewTokenService(iTokenDao, iTokenValidityPolicyDao)
)

func generateKey() ([]byte, error) {
	key := make([]byte, 16)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func TestGenerateAPIKey(t *testing.T) {
	token, err := tokenService.GenerateToken()
	if err != nil {
		t.Fatalf("Error getting token: %v", err)
	}
	// 验证 AccessToken 是否非空
	if token == "" {
		t.Fatal("token is empty")
	}
	t.Logf("Token: %s", token)
}

func TestGenerateAPIKeyWithKey(t *testing.T) {
	key, err := generateKey()
	if err != nil {
		t.Fatalf("Error generating key: %v", err)
	}
	t.Logf("Generated key: %x\n", key)

	ID := dao.NewUUID()
	t.Logf("uuid: %s", ID)

	token := models.TokenResp{
		Token:           "test-000001",
		AppScenarioName: "test",
	}
	// 将token转换为JSON
	jsonData, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("Error marshalling token: %v", err)
	}
	// 加密JSON数据
	encryptedData, err := utils.EncryptData(jsonData)
	if err != nil {
		t.Fatalf("Error encrypting data: %v", err)
	}
	t.Logf("encryptedData: %x", encryptedData)

	// 解密数据
	decryptedData, err := utils.DecryptData(encryptedData)
	if err != nil {
		t.Fatalf("Error decrypting data: %v", err)
	}
	t.Logf("decryptedData: %s", decryptedData)
	// 将解密后的数据转换为TokenResp结构体
	var decryptedToken models.TokenResp
	err = json.Unmarshal(decryptedData, &decryptedToken)
	if err != nil {
		t.Fatalf("Error unmarshalling data: %v", err)
	}
	t.Logf("decryptedToken: %+v", decryptedToken)
}
