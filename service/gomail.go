package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"gopkg.in/gomail.v2"
)

// 定义邮箱服务器连接信息，qq邮箱填授权码
type MailConfig struct {
	Host string `json:"Host"` //邮箱服务器地址 smtp.qq.com
	Port int    `json:"Port"` //邮箱服务器端口 smtp.qq.com 端口465
	User string `json:"User"` //邮箱账号 1971897775@qq.com
	Pass string `json:"Pass"` //邮箱密码  btizzlfqnlidefbi
	SSL  bool   `json:"SSL"`  //SSL加密传输，端口==465是默认为true
}

func NewMailConfig() *MailConfig {
	return &MailConfig{
		Host: "smtp.qq.com",
		Port: 465,
		User: "1971897775@qq.com",
		Pass: "btizzlfqnlidefbi",
		SSL:  true,
	}
}

// mailConfig 邮件配置MailConfig 不能为空
// mailFrom 邮件发送方 默认MailConfig账号
// mailTo 接收邮箱 不能为空
// subject 主题
// message 发送消息
func SendMail(mailFrom string, mailTo string) (string, error) {
	mailConfig := NewMailConfig()
	//接收消息的邮箱不能为空
	if mailTo == "" {
		return "", errors.New("mailTo:接收消息的邮箱不能为空")
	}

	code := GenerateCode()
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(mailConfig.User, mailFrom)) //这种方式可以添加别名，
	m.SetHeader("To", mailTo)                                       //发送给用户
	m.SetHeader("Subject", "验证码")                                   //设置邮件主题
	m.SetBody("text/html", retHTMLMessage(mailTo, code))            //设置邮件正文

	//连接
	dialer := gomail.NewDialer(mailConfig.Host, mailConfig.Port, mailConfig.User, mailConfig.Pass)

	//发送
	erBySend := dialer.DialAndSend(m)
	if erBySend != nil {
		fmt.Printf("base package SendMail send error:%v", erBySend.Error())
		//error
		return "", erBySend
	}
	//success
	return code, nil
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
