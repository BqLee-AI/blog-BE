package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database config
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// Mail config
	MailHost     string
	MailPort     int
	MailUsername string
	MailPassword string

	// App config
	AppPort int
	GinMode string
}

var AppConfig *Config

// LoadConfig 从 .env 文件加载配置；如果文件不存在，则继续使用进程环境变量和默认值。
func LoadConfig(envPath string) error {
	if envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err != nil {
				return fmt.Errorf("failed to load .env file: %v", err)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check .env file: %v", err)
		}
	}

	AppConfig = &Config{
		// Database config
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "admin"),
		DBPassword: getEnv("DB_PASSWORD", "123456"),
		DBName:     getEnv("DB_NAME", "mydb"),

		// Mail config
		MailHost:     getEnv("MAIL_HOST", "smtp.qq.com"),
		MailPort:     getEnvAsInt("MAIL_PORT", 465),
		MailUsername: getEnv("MAIL_USERNAME", ""),
		MailPassword: getEnv("MAIL_PASSWORD", ""),

		// App config
		AppPort: getEnvAsInt("APP_PORT", 8080),
		GinMode: getEnv("GIN_MODE", "debug"),
	}

	return nil
}

// getEnv 获取环境变量，如果不存在返回默认值
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvAsInt 获取环境变量并转换为整数，如果不存在返回默认值
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}
