package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Mail     MailConfig     `mapstructure:"mail"`
	CORS     CORSConfig     `mapstructure:"cors"`
}

type ServerConfig struct {
	Port           int      `mapstructure:"port"`
	Mode           string   `mapstructure:"mode"`
	TrustedProxies []string `mapstructure:"trusted_proxies"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"ssl_mode"`
	TimeZone string `mapstructure:"time_zone"`
}

type JWTConfig struct {
	PrivateKeyPath string        `mapstructure:"private_key_path"`
	PublicKeyPath  string        `mapstructure:"public_key_path"`
	PrivateKey     string        `mapstructure:"private_key"`
	PublicKey      string        `mapstructure:"public_key"`
	AccessExpire   time.Duration `mapstructure:"access_expire"`
	RefreshExpire  time.Duration `mapstructure:"refresh_expire"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MailConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

var (
	AppConfig *Config

	currentConfig atomic.Pointer[Config]
	watcherOnce   sync.Once
)

func init() {
	setCurrentConfig(defaultConfig())
}

func LoadConfig() error {
	files, err := resolveConfigFiles()
	if err != nil {
		return err
	}

	if err := loadConfigFiles(files, false); err != nil {
		return err
	}

	var watchErr error
	watcherOnce.Do(func() {
		watchErr = startConfigWatcher(files)
	})

	return watchErr
}

func Get() *Config {
	if cfg := currentConfig.Load(); cfg != nil {
		return cfg
	}

	defaultCfg := defaultConfig()
	setCurrentConfig(defaultCfg)
	return currentConfig.Load()
}

func loadConfigFromFile(filePath string, watch bool) error {
	if filePath == "" {
		return fmt.Errorf("config file path is empty")
	}

	if err := loadConfigFiles([]string{filePath}, false); err != nil {
		return err
	}

	if watch {
		var watchErr error
		watcherOnce.Do(func() {
			watchErr = startConfigWatcher([]string{filePath})
		})
		return watchErr
	}

	return nil
}

func loadConfigFiles(files []string, watch bool) error {
	v, err := buildViper(files)
	if err != nil {
		return err
	}

	if err := applyConfig(v); err != nil {
		return err
	}

	if watch {
		var watchErr error
		watcherOnce.Do(func() {
			watchErr = startConfigWatcher(files)
		})
		return watchErr
	}

	return nil
}

func resolveConfigFiles() ([]string, error) {
	if explicit := strings.TrimSpace(os.Getenv("APP_CONFIG_FILE")); explicit != "" {
		return []string{explicit}, nil
	}

	env := strings.TrimSpace(os.Getenv("APP_ENV"))
	if env == "" {
		env = "development"
	}

	searchPaths := []string{".", "./config"}
	primaryCandidates := []string{
		fmt.Sprintf("config.%s.yaml", env),
		fmt.Sprintf("config.%s.yml", env),
		fmt.Sprintf("config.%s.json", env),
		fmt.Sprintf("config.%s.env", env),
		"config.yaml",
		"config.yml",
		"config.json",
		"config.env",
	}

	primaryFile, err := findFirstExistingFile(searchPaths, primaryCandidates)
	if err != nil {
		return nil, err
	}

	if primaryFile == "" {
		fallbackFile, fallbackErr := findFirstExistingFile(searchPaths, []string{fmt.Sprintf(".env.%s", env), ".env"})
		if fallbackErr != nil {
			return nil, fallbackErr
		}
		if fallbackFile == "" {
			return nil, fmt.Errorf("no config file found for APP_ENV=%q", env)
		}
		return []string{fallbackFile}, nil
	}

	files := []string{primaryFile}
	for _, candidate := range []string{fmt.Sprintf(".env.%s", env), ".env"} {
		overrideFile, overrideErr := findFirstExistingFile(searchPaths, []string{candidate})
		if overrideErr != nil {
			return nil, overrideErr
		}
		if overrideFile != "" && !samePath(overrideFile, primaryFile) {
			files = append(files, overrideFile)
		}
	}

	return files, nil
}

func buildViper(files []string) (*viper.Viper, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no config files provided")
	}

	v := viper.New()
	configureViper(v)

	for idx, file := range files {
		setConfigFile(v, file)

		var err error
		if idx == 0 {
			err = v.ReadInConfig()
		} else {
			err = v.MergeInConfig()
		}
		if err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", file, err)
		}
	}

	applyLegacyKeyMappings(v)
	return v, nil
}

func configureViper(v *viper.Viper) {
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)
	bindEnv(v, "server.port", "SERVER_PORT", "APP_PORT")
	bindEnv(v, "server.mode", "SERVER_MODE", "GIN_MODE")
	bindEnv(v, "server.trusted_proxies", "SERVER_TRUSTED_PROXIES", "GIN_TRUSTED_PROXIES")

	bindEnv(v, "database.host", "DATABASE_HOST", "DB_HOST")
	bindEnv(v, "database.port", "DATABASE_PORT", "DB_PORT")
	bindEnv(v, "database.user", "DATABASE_USER", "DB_USER")
	bindEnv(v, "database.password", "DATABASE_PASSWORD", "DB_PASSWORD")
	bindEnv(v, "database.name", "DATABASE_NAME", "DB_NAME")
	bindEnv(v, "database.ssl_mode", "DATABASE_SSL_MODE")
	bindEnv(v, "database.time_zone", "DATABASE_TIME_ZONE")

	bindEnv(v, "redis.addr", "REDIS_ADDR")
	bindEnv(v, "redis.password", "REDIS_PASSWORD")
	bindEnv(v, "redis.db", "REDIS_DB")

	bindEnv(v, "mail.host", "MAIL_HOST")
	bindEnv(v, "mail.port", "MAIL_PORT")
	bindEnv(v, "mail.username", "MAIL_USERNAME")
	bindEnv(v, "mail.password", "MAIL_PASSWORD")

	bindEnv(v, "jwt.private_key_path", "JWT_PRIVATE_KEY_PATH")
	bindEnv(v, "jwt.public_key_path", "JWT_PUBLIC_KEY_PATH")
	bindEnv(v, "jwt.private_key", "JWT_PRIVATE_KEY")
	bindEnv(v, "jwt.public_key", "JWT_PUBLIC_KEY")
	bindEnv(v, "jwt.access_expire", "JWT_ACCESS_EXPIRE", "JWT_ACCESS_TTL")
	bindEnv(v, "jwt.refresh_expire", "JWT_REFRESH_EXPIRE", "JWT_REFRESH_TTL")

	bindEnv(v, "cors.allow_origins", "CORS_ALLOW_ORIGINS")
	bindEnv(v, "cors.allow_methods", "CORS_ALLOW_METHODS")
	bindEnv(v, "cors.allow_headers", "CORS_ALLOW_HEADERS")
	bindEnv(v, "cors.expose_headers", "CORS_EXPOSE_HEADERS")
	bindEnv(v, "cors.allow_credentials", "CORS_ALLOW_CREDENTIALS")
}

func applyConfig(v *viper.Viper) error {
	nextConfig := defaultConfig()
	if err := v.Unmarshal(&nextConfig, func(dc *mapstructure.DecoderConfig) {
		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			stringToSliceHook(),
			flexibleDurationHook(),
		)
	}); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	setCurrentConfig(nextConfig)
	log.Printf("config loaded from %s", v.ConfigFileUsed())
	return nil
}

func startConfigWatcher(files []string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create config watcher: %w", err)
	}

	fileSet := make(map[string]struct{}, len(files))
	dirSet := make(map[string]struct{})
	for _, file := range files {
		absFile, absErr := filepath.Abs(file)
		if absErr != nil {
			_ = watcher.Close()
			return fmt.Errorf("failed to resolve config path %s: %w", file, absErr)
		}

		cleanFile := filepath.Clean(absFile)
		fileSet[cleanFile] = struct{}{}

		dir := filepath.Dir(cleanFile)
		if _, exists := dirSet[dir]; exists {
			continue
		}
		if addErr := watcher.Add(dir); addErr != nil {
			_ = watcher.Close()
			return fmt.Errorf("failed to watch config directory %s: %w", dir, addErr)
		}
		dirSet[dir] = struct{}{}
	}

	go func() {
		defer watcher.Close()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
					continue
				}

				if !matchesWatchedFile(event.Name, fileSet) {
					continue
				}

				if err := loadConfigFiles(files, false); err != nil {
					log.Printf("config reload failed: %v", err)
					continue
				}
				log.Printf("config reloaded: %s", event.Name)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("config watch error: %v", err)
			}
		}
	}()

	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.trusted_proxies", []string{"127.0.0.1", "::1"})

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "admin")
	v.SetDefault("database.password", "")
	v.SetDefault("database.name", "mydb")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.time_zone", "Asia/Shanghai")

	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	v.SetDefault("mail.host", "smtp.qq.com")
	v.SetDefault("mail.port", 465)
	v.SetDefault("mail.username", "")
	v.SetDefault("mail.password", "")

	v.SetDefault("jwt.private_key_path", "./secrets/jwt_private.pem")
	v.SetDefault("jwt.public_key_path", "./secrets/jwt_public.pem")
	v.SetDefault("jwt.private_key", "")
	v.SetDefault("jwt.public_key", "")
	v.SetDefault("jwt.access_expire", "15m")
	v.SetDefault("jwt.refresh_expire", "168h")

	v.SetDefault("cors.allow_origins", []string{"http://localhost:5173", "http://127.0.0.1:5173"})
	v.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"})
	v.SetDefault("cors.expose_headers", []string{"Content-Length"})
	v.SetDefault("cors.allow_credentials", false)
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Port:           8080,
			Mode:           "debug",
			TrustedProxies: []string{"127.0.0.1", "::1"},
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "admin",
			Password: "",
			Name:     "mydb",
			SSLMode:  "disable",
			TimeZone: "Asia/Shanghai",
		},
		JWT: JWTConfig{
			PrivateKeyPath: "./secrets/jwt_private.pem",
			PublicKeyPath:  "./secrets/jwt_public.pem",
			AccessExpire:   15 * time.Minute,
			RefreshExpire:  7 * 24 * time.Hour,
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Mail: MailConfig{
			Host:     "smtp.qq.com",
			Port:     465,
			Username: "",
			Password: "",
		},
		CORS: CORSConfig{
			AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: false,
		},
	}
}

func setCurrentConfig(cfg Config) {
	config := cfg
	currentConfig.Store(&config)
	AppConfig = &config
}

func setConfigFile(v *viper.Viper, filePath string) {
	v.SetConfigFile(filePath)
	if ext := strings.TrimPrefix(filepath.Ext(filePath), "."); ext != "" {
		v.SetConfigType(ext)
	}
}

func bindEnv(v *viper.Viper, key string, envNames ...string) {
	args := append([]string{key}, envNames...)
	if err := v.BindEnv(args...); err != nil {
		panic(err)
	}
}

func applyLegacyKeyMappings(v *viper.Viper) {
	legacyKeys := map[string][]string{
		"server.port":            {"SERVER_PORT", "APP_PORT"},
		"server.mode":            {"SERVER_MODE", "GIN_MODE"},
		"server.trusted_proxies": {"SERVER_TRUSTED_PROXIES", "GIN_TRUSTED_PROXIES"},
		"database.host":          {"DATABASE_HOST", "DB_HOST"},
		"database.port":          {"DATABASE_PORT", "DB_PORT"},
		"database.user":          {"DATABASE_USER", "DB_USER"},
		"database.password":      {"DATABASE_PASSWORD", "DB_PASSWORD"},
		"database.name":          {"DATABASE_NAME", "DB_NAME"},
		"database.ssl_mode":      {"DATABASE_SSL_MODE"},
		"database.time_zone":     {"DATABASE_TIME_ZONE"},
		"redis.addr":             {"REDIS_ADDR"},
		"redis.password":         {"REDIS_PASSWORD"},
		"redis.db":               {"REDIS_DB"},
		"mail.host":              {"MAIL_HOST"},
		"mail.port":              {"MAIL_PORT"},
		"mail.username":          {"MAIL_USERNAME"},
		"mail.password":          {"MAIL_PASSWORD"},
		"jwt.private_key_path":   {"JWT_PRIVATE_KEY_PATH"},
		"jwt.public_key_path":    {"JWT_PUBLIC_KEY_PATH"},
		"jwt.private_key":        {"JWT_PRIVATE_KEY"},
		"jwt.public_key":         {"JWT_PUBLIC_KEY"},
		"jwt.access_expire":      {"JWT_ACCESS_EXPIRE", "JWT_ACCESS_TTL"},
		"jwt.refresh_expire":     {"JWT_REFRESH_EXPIRE", "JWT_REFRESH_TTL"},
		"cors.allow_origins":     {"CORS_ALLOW_ORIGINS"},
		"cors.allow_methods":     {"CORS_ALLOW_METHODS"},
		"cors.allow_headers":     {"CORS_ALLOW_HEADERS"},
		"cors.expose_headers":    {"CORS_EXPOSE_HEADERS"},
		"cors.allow_credentials": {"CORS_ALLOW_CREDENTIALS"},
	}

	for targetKey, sourceKeys := range legacyKeys {
		for _, sourceKey := range sourceKeys {
			if !v.IsSet(sourceKey) {
				continue
			}
			v.Set(targetKey, v.Get(sourceKey))
			break
		}
	}
}

func findFirstExistingFile(searchPaths []string, candidates []string) (string, error) {
	for _, searchPath := range searchPaths {
		for _, candidate := range candidates {
			fullPath := filepath.Join(searchPath, candidate)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath, nil
			} else if !os.IsNotExist(err) {
				return "", fmt.Errorf("failed to inspect %s: %w", fullPath, err)
			}
		}
	}

	return "", nil
}

func matchesWatchedFile(eventPath string, watchedFiles map[string]struct{}) bool {
	absPath, err := filepath.Abs(eventPath)
	if err != nil {
		return false
	}

	cleanPath := filepath.Clean(absPath)
	for watchedFile := range watchedFiles {
		if samePath(cleanPath, watchedFile) {
			return true
		}
	}

	return false
}

func samePath(left, right string) bool {
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}

func stringToSliceHook() mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() != reflect.String || to != reflect.TypeOf([]string{}) {
			return data, nil
		}

		value := strings.TrimSpace(data.(string))
		if value == "" {
			return []string{}, nil
		}

		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}

		return result, nil
	}
}

func flexibleDurationHook() mapstructure.DecodeHookFuncType {
	durationType := reflect.TypeOf(time.Duration(0))

	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if to != durationType {
			return data, nil
		}

		switch value := data.(type) {
		case string:
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				return time.Duration(0), nil
			}

			if duration, err := time.ParseDuration(trimmed); err == nil {
				return duration, nil
			}

			seconds, err := strconv.ParseInt(trimmed, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid duration value %q", value)
			}
			return time.Duration(seconds) * time.Second, nil
		case int:
			return time.Duration(value) * time.Second, nil
		case int64:
			return time.Duration(value) * time.Second, nil
		case float64:
			return time.Duration(value) * time.Second, nil
		default:
			return data, nil
		}
	}
}
