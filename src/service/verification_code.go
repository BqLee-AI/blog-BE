package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrVerificationCodeNotFound = errors.New("verification code not found")
	ErrVerificationCodeExpired  = errors.New("verification code expired")
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

type verificationCodeRecord struct {
	Code          string
	ExpiresAt     time.Time
	CooldownUntil time.Time
}

var verificationCodeStore = struct {
	sync.Mutex
	items map[string]verificationCodeRecord
}{
	items: make(map[string]verificationCodeRecord),
}

const (
	verificationCodeTTL      = 10 * time.Minute
	verificationCodeCooldown = 1 * time.Minute
)

func normalizeVerificationEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func SendVerificationCode(mailTo string) error {
	normalizedEmail := normalizeVerificationEmail(mailTo)
	if normalizedEmail == "" {
		return errors.New("mailTo:接收消息的邮箱不能为空")
	}

	now := time.Now()
	if remaining, limited := getVerificationCodeCooldown(normalizedEmail, now); limited {
		return &VerificationCooldownError{Remaining: remaining}
	}

	code := GenerateCode()
	reserveVerificationCode(normalizedEmail, now.Add(verificationCodeCooldown))

	if err := SendMail("", mailTo, code); err != nil {
		clearVerificationCode(normalizedEmail)
		return err
	}

	storeVerificationCode(normalizedEmail, code, now)
	return nil
}

func VerifyVerificationCode(mailTo string, code string) error {
	normalizedEmail := normalizeVerificationEmail(mailTo)
	if normalizedEmail == "" {
		return ErrVerificationCodeNotFound
	}

	verificationCodeStore.Lock()
	defer verificationCodeStore.Unlock()

	record, exists := verificationCodeStore.items[normalizedEmail]
	if !exists {
		return ErrVerificationCodeNotFound
	}

	now := time.Now()
	if now.After(record.ExpiresAt) {
		delete(verificationCodeStore.items, normalizedEmail)
		return ErrVerificationCodeExpired
	}

	if strings.TrimSpace(code) == "" || record.Code != strings.TrimSpace(code) {
		return ErrVerificationCodeInvalid
	}

	delete(verificationCodeStore.items, normalizedEmail)
	return nil
}

func getVerificationCodeCooldown(email string, now time.Time) (time.Duration, bool) {
	verificationCodeStore.Lock()
	defer verificationCodeStore.Unlock()

	record, exists := verificationCodeStore.items[email]
	if !exists {
		return 0, false
	}

	if now.After(record.ExpiresAt) {
		delete(verificationCodeStore.items, email)
		return 0, false
	}

	if now.Before(record.CooldownUntil) {
		return record.CooldownUntil.Sub(now), true
	}

	return 0, false
}

func reserveVerificationCode(email string, cooldownUntil time.Time) {
	verificationCodeStore.Lock()
	defer verificationCodeStore.Unlock()

	verificationCodeStore.items[email] = verificationCodeRecord{
		ExpiresAt:     cooldownUntil,
		CooldownUntil: cooldownUntil,
	}
}

func storeVerificationCode(email string, code string, now time.Time) {
	verificationCodeStore.Lock()
	defer verificationCodeStore.Unlock()

	verificationCodeStore.items[email] = verificationCodeRecord{
		Code:          code,
		ExpiresAt:     now.Add(verificationCodeTTL),
		CooldownUntil: now.Add(verificationCodeCooldown),
	}
}

func clearVerificationCode(email string) {
	verificationCodeStore.Lock()
	defer verificationCodeStore.Unlock()

	delete(verificationCodeStore.items, email)
}
