package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"blog-BE/src/config"

	"github.com/redis/go-redis/v9"
)

var (
	ErrVerificationCodeNotFound = errors.New("verification code not found")
	ErrVerificationCodeInvalid  = errors.New("verification code invalid")
)

type VerificationCooldownError struct {
	Remaining time.Duration
}

func (e *VerificationCooldownError) Error() string {
	seconds := int(e.Remaining.Seconds())
	if e.Remaining%time.Second != 0 {
		seconds++
	}
	if seconds < 1 {
		seconds = 1
	}
	return fmt.Sprintf("请在 %d 秒后重新发送验证码", seconds)
}

const (
	verificationCodeTTL      = 5 * time.Minute
	verificationCodeCooldown = 1 * time.Minute
)

func normalizeVerificationEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func verificationCodeKey(email string) string {
	return fmt.Sprintf("verify:%s", email)
}

func verificationCooldownKey(email string) string {
	return fmt.Sprintf("verify:cooldown:%s", email)
}

func SendVerificationCode(mailTo string) error {
	normalizedEmail := normalizeVerificationEmail(mailTo)
	if normalizedEmail == "" {
		return errors.New("mailTo:接收消息的邮箱不能为空")
	}

	if config.RedisClient == nil {
		return errors.New("redis client is not initialized")
	}

	ctx := context.Background()
	cooldownKey := verificationCooldownKey(normalizedEmail)
	codeKey := verificationCodeKey(normalizedEmail)

	if remaining, limited, err := getVerificationCooldown(ctx, cooldownKey); err != nil {
		return err
	} else if limited {
		return &VerificationCooldownError{Remaining: remaining}
	}

	code := GenerateCode()

	if err := config.RedisClient.Set(ctx, cooldownKey, "1", verificationCodeCooldown).Err(); err != nil {
		return err
	}
	if err := config.RedisClient.Set(ctx, codeKey, code, verificationCodeTTL).Err(); err != nil {
		_ = config.RedisClient.Del(ctx, cooldownKey).Err()
		return err
	}

	if err := SendMail("", mailTo, code); err != nil {
		_, _ = config.RedisClient.Del(ctx, cooldownKey, codeKey).Result()
		return err
	}

	return nil
}

func VerifyCode(email, code string) bool {
	normalizedEmail := normalizeVerificationEmail(email)
	if normalizedEmail == "" || strings.TrimSpace(code) == "" {
		return false
	}

	if config.RedisClient == nil {
		return false
	}

	ctx := context.Background()
	storedCode, err := config.RedisClient.Get(ctx, verificationCodeKey(normalizedEmail)).Result()
	if err != nil {
		return false
	}

	return storedCode == strings.TrimSpace(code)
}

func VerifyVerificationCode(mailTo string, code string) error {
	normalizedEmail := normalizeVerificationEmail(mailTo)
	if normalizedEmail == "" {
		return ErrVerificationCodeNotFound
	}

	if config.RedisClient == nil {
		return errors.New("redis client is not initialized")
	}

	ctx := context.Background()
	storedCode, err := config.RedisClient.Get(ctx, verificationCodeKey(normalizedEmail)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrVerificationCodeNotFound
		}
		return err
	}

	if storedCode != strings.TrimSpace(code) {
		return ErrVerificationCodeInvalid
	}

	if err := config.RedisClient.Del(ctx, verificationCodeKey(normalizedEmail)).Err(); err != nil {
		return err
	}

	return nil
}

func getVerificationCooldown(ctx context.Context, cooldownKey string) (time.Duration, bool, error) {
	ttl, err := config.RedisClient.TTL(ctx, cooldownKey).Result()
	if err != nil {
		return 0, false, err
	}

	if ttl > 0 {
		return ttl, true, nil
	}

	return 0, false, nil
}
