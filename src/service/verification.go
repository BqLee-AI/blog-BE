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

var verifyAndConsumeCodeScript = redis.NewScript(`
local current = redis.call("GET", KEYS[1])
if not current then
  return 0
end
if current ~= ARGV[1] then
  return -1
end
redis.call("DEL", KEYS[1])
return 1
`)

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

	ok, err := config.RedisClient.SetNX(ctx, cooldownKey, "1", verificationCodeCooldown).Result()
	if err != nil {
		return err
	}
	if !ok {
		ttl, ttlErr := config.RedisClient.TTL(ctx, cooldownKey).Result()
		if ttlErr != nil {
			return ttlErr
		}
		if ttl <= 0 {
			ttl = verificationCodeCooldown
		}
		return &VerificationCooldownError{Remaining: ttl}
	}

	code := GenerateCode()

	if err := config.RedisClient.Set(ctx, codeKey, code, verificationCodeTTL).Err(); err != nil {
		_ = config.RedisClient.Del(ctx, cooldownKey).Err()
		return err
	}

	if err := SendMail("", normalizedEmail, code); err != nil {
		_, _ = config.RedisClient.Del(ctx, cooldownKey, codeKey).Result()
		return err
	}

	return nil
}

func VerifyCode(email, code string) bool {
	return VerifyVerificationCode(email, code) == nil
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
	result, err := verifyAndConsumeCodeScript.Run(
		ctx,
		config.RedisClient,
		[]string{verificationCodeKey(normalizedEmail)},
		strings.TrimSpace(code),
	).Int()
	if err != nil {
		return err
	}

	switch result {
	case 1:
		return nil
	case 0:
		return ErrVerificationCodeNotFound
	default:
		return ErrVerificationCodeInvalid
	}
}
