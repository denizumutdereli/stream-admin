package caesar

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

type OTP struct {
	Code   string    `json:"code"`
	Expiry time.Time `json:"expiry"`
}

type CaesarManager interface {
	GenerateOTP() (string, error)
	StoreOTP(phone, otp string, expiry time.Time) error
	RetrieveOTP(phone string) (string, error)
	Generate2FACaesar() (string, error)
}

type caesarManager struct {
	redisClient *transport.RedisManager
}

func NewCaesarManager(redisClient *transport.RedisManager) CaesarManager {
	return &caesarManager{
		redisClient: redisClient,
	}
}

func (sm *caesarManager) GenerateOTP() (string, error) {
	otpBytes := make([]byte, 3)
	if _, err := rand.Read(otpBytes); err != nil {
		return "", err
	}

	otp := int32(otpBytes[0])<<16 | int32(otpBytes[1])<<8 | int32(otpBytes[2])
	otp %= 1000000

	return fmt.Sprintf("%06d", otp), nil
}

func (c *caesarManager) StoreOTP(phone, otp string, expiry time.Time) error {
	ctx := context.Background()

	otpCode := &OTP{
		Expiry: expiry,
		Code:   otp,
	}

	marshalledCode, err := json.Marshal(otpCode)
	if err != nil {
		return fmt.Errorf("failed to marshal OTP: %v", err)
	}

	ttl := time.Until(expiry)
	if err := c.redisClient.SetKeyValue(ctx, phone, marshalledCode, ttl); err != nil {
		return err
	}

	return nil
}

func (c *caesarManager) RetrieveOTP(phone string) (string, error) {
	ctx := context.Background()
	var result []byte

	err := c.redisClient.GetKeyValue(ctx, phone, &result)
	if err == redis.Nil {
		return "", errors.New("no OTP code found")
	} else if err != nil {
		return "", err
	}

	var otpCode OTP
	if err = json.Unmarshal([]byte(result), &otpCode); err != nil {
		return "", fmt.Errorf("error unmarshaling OTP code: %v", err)
	}

	if time.Now().After(otpCode.Expiry) {
		return "", errors.New("OTP code has expired")
	}

	return otpCode.Code, nil
}

func (c *caesarManager) Generate2FACaesar() (string, error) {
	randomKey := make([]byte, 20)
	if _, err := rand.Read(randomKey); err != nil {
		return "", err
	}

	secretKey := base32.StdEncoding.EncodeToString(randomKey)
	return secretKey, nil
}
