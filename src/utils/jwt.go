package utils

import (
	"blog-BE/src/config"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	defaultAccessTTL  = 15 * time.Minute
	defaultRefreshTTL = 7 * 24 * time.Hour
)

type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type Claims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	RoleID    uint   `json:"role_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

var (
	keyMu            sync.RWMutex
	privateKey       *rsa.PrivateKey
	publicKey        *rsa.PublicKey
	loadKeyError     error
	loadedKeyVersion string
	keysLoaded       bool
)

func GenerateTokenPair(userID uint, username string, roleID uint) (*TokenPair, error) {
	privKey, _, err := loadRSAKeyPair()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	accessExpiresAt := now.Add(getAccessTTL())
	refreshExpiresAt := now.Add(getRefreshTTL())

	accessToken, err := signToken(privKey, Claims{
		UserID:    userID,
		Username:  username,
		RoleID:    roleID,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
		},
	})
	if err != nil {
		return nil, err
	}

	refreshToken, err := signToken(privKey, Claims{
		UserID:    userID,
		Username:  username,
		RoleID:    roleID,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
		},
	})
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func RefreshTokenPair(refreshToken string) (*TokenPair, error) {
	claims, err := ParseToken(refreshToken, TokenTypeRefresh)
	if err != nil {
		return nil, err
	}

	return GenerateTokenPair(claims.UserID, claims.Username, claims.RoleID)
}

func ParseAccessToken(tokenString string) (*Claims, error) {
	return ParseToken(tokenString, TokenTypeAccess)
}

func ParseToken(tokenString string, expectedType string) (*Claims, error) {
	_, pubKey, err := loadRSAKeyPair()
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("unexpected token type: %s", claims.TokenType)
	}

	return claims, nil
}

func ExtractBearerToken(headerValue string) string {
	if headerValue == "" {
		return ""
	}
	parts := strings.Fields(headerValue)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

func loadRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	jwtConfig := config.Get().JWT
	version := strings.Join([]string{
		jwtConfig.PrivateKeyPath,
		jwtConfig.PublicKeyPath,
		jwtConfig.PrivateKey,
		jwtConfig.PublicKey,
	}, "\x00")

	keyMu.RLock()
	if keysLoaded && loadedKeyVersion == version {
		defer keyMu.RUnlock()
		return privateKey, publicKey, loadKeyError
	}
	keyMu.RUnlock()

	keyMu.Lock()
	defer keyMu.Unlock()

	if keysLoaded && loadedKeyVersion == version {
		return privateKey, publicKey, loadKeyError
	}

	privateKey, publicKey, loadKeyError = loadKeys(jwtConfig)
	loadedKeyVersion = version
	keysLoaded = true

	return privateKey, publicKey, loadKeyError
}

func loadKeys(jwtConfig config.JWTConfig) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privatePEM, err := getKeyMaterial(jwtConfig.PrivateKeyPath, jwtConfig.PrivateKey)
	if err != nil {
		return nil, nil, err
	}

	privateBlock, _ := pem.Decode([]byte(privatePEM))
	if privateBlock == nil {
		return nil, nil, errors.New("failed to decode JWT private key PEM")
	}

	parsedPrivate, err := parseRSAPrivateKey(privateBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse JWT private key: %w", err)
	}

	publicPEM, publicErr := getOptionalKeyMaterial(jwtConfig.PublicKeyPath, jwtConfig.PublicKey)
	if publicErr != nil {
		return nil, nil, publicErr
	}

	parsedPublic := &parsedPrivate.PublicKey
	if publicPEM != "" {
		publicBlock, _ := pem.Decode([]byte(publicPEM))
		if publicBlock == nil {
			return nil, nil, errors.New("failed to decode JWT public key PEM")
		}

		parsedPublic, err = parseRSAPublicKey(publicBlock.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse JWT public key: %w", err)
		}
	}

	return parsedPrivate, parsedPublic, nil
}

func getKeyMaterial(pathKey, valueKey string) (string, error) {
	if path := strings.TrimSpace(pathKey); path != "" {
		content, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read key file %s: %w", path, err)
		}
		return strings.TrimSpace(string(content)), nil
	}

	value := strings.TrimSpace(valueKey)
	if value == "" {
		return "", fmt.Errorf("missing JWT key material")
	}

	return normalizePEM(value), nil
}

func getOptionalKeyMaterial(pathKey, valueKey string) (string, error) {
	if path := strings.TrimSpace(pathKey); path != "" {
		content, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read key file %s: %w", path, err)
		}
		return strings.TrimSpace(string(content)), nil
	}

	value := strings.TrimSpace(valueKey)
	if value == "" {
		return "", nil
	}

	return normalizePEM(value), nil
}

func normalizePEM(value string) string {
	return strings.ReplaceAll(value, "\\n", "\n")
}

func parseRSAPrivateKey(keyBytes []byte) (*rsa.PrivateKey, error) {
	if parsed, err := x509.ParsePKCS1PrivateKey(keyBytes); err == nil {
		return parsed, nil
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}

	privateKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("JWT private key is not an RSA private key")
	}

	return privateKey, nil
}

func parseRSAPublicKey(keyBytes []byte) (*rsa.PublicKey, error) {
	if parsed, err := x509.ParsePKIXPublicKey(keyBytes); err == nil {
		publicKey, ok := parsed.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("JWT public key is not an RSA public key")
		}
		return publicKey, nil
	}

	parsedKey, err := x509.ParsePKCS1PublicKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return parsedKey, nil
}

func signToken(privKey *rsa.PrivateKey, claims Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privKey)
}

func getAccessTTL() time.Duration {
	ttl := config.Get().JWT.AccessExpire
	if ttl <= 0 {
		return defaultAccessTTL
	}
	return ttl
}

func getRefreshTTL() time.Duration {
	ttl := config.Get().JWT.RefreshExpire
	if ttl <= 0 {
		return defaultRefreshTTL
	}
	return ttl
}
