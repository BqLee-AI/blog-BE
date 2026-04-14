package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"blog-BE/src/config"

	"gopkg.in/gomail.v2"
)

func SendMail(mailFrom string, mailTo string, code string) error {
	mailConfig := config.Get().Mail

	// 接收消息的邮箱不能为空
	if mailTo == "" {
		return errors.New("mailTo:接收消息的邮箱不能为空")
	}

	fromAddress := mailConfig.Username
	if mailFrom != "" {
		fromAddress = mailFrom
	}
	if fromAddress == "" {
		return errors.New("mailFrom:发件人邮箱不能为空")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", fromAddress)
	m.SetHeader("To", mailTo)                            // 发送给用户
	m.SetHeader("Subject", "验证码")                        // 设置邮件主题
	m.SetBody("text/html", retHTMLMessage(mailTo, code)) // 设置邮件正文

	// 连接
	dialer := gomail.NewDialer(mailConfig.Host, mailConfig.Port, mailConfig.Username, mailConfig.Password)

	// 发送
	sendResult := make(chan error, 1)
	go func() {
		sendResult <- dialer.DialAndSend(m)
	}()

	select {
	case erBySend := <-sendResult:
		if erBySend != nil {
			// error
			return erBySend
		}
	case <-time.After(time.Minute):
		return errors.New("send verification code timeout after 1 minute")
	}
	// success
	return nil
}

// 生成随机验证码
func GenerateCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}

func retHTMLMessage(email string, code string) string {
	return fmt.Sprintf(`<h2>您好，%s</h2>
						<p>您的验证码为：<strong>%s</strong></p>
						<p>验证码将于 <strong>%s</strong> 后失效，请勿泄露给他人。</p>
						<p>如非本人操作，请忽略此邮件。</p>
						<p>此邮件由系统自动发送，请勿回复</p>
						<p>© 2026 BqLee Blog</p>`, email, code, "60s")
}
