package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all the configuration for the application
type Config struct {
	App    AppConfig    `mapstructure:"app"`
	Server ServerConfig `mapstructure:"server"`
	API    APIConfig    `mapstructure:"api"`
	Auth   AuthConfig   `mapstructure:"auth"`
	Log    LogConfig    `mapstructure:"log"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Description string `mapstructure:"description"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Env             string        `mapstructure:"env"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes"`
	MaxConnections  int           `mapstructure:"max_connections"`
	MaxIdleConns    int           `mapstructure:"max_idle_connections"`
	RequestTimeout  string        `mapstructure:"request_timeout"`
	MaxRequestSize  int           `mapstructure:"max_request_size"`
	MaxResponseSize int           `mapstructure:"max_response_size"`
}

// APIConfig holds API-specific configuration
type APIConfig struct {
	Version   string          `mapstructure:"version"`
	Prefix    string          `mapstructure:"prefix"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	CORS      CORSConfig      `mapstructure:"cors"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	MaxRequests int    `mapstructure:"max_requests"`
	Window      string `mapstructure:"window"`
	Burst       int    `mapstructure:"burst"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	AllowOrigin      string `mapstructure:"allow_origin"`
	AllowMethods     string `mapstructure:"allow_methods"`
	AllowHeaders     string `mapstructure:"allow_headers"`
	ExposeHeaders    string `mapstructure:"expose_headers"`
	MaxAge           int    `mapstructure:"max_age"`
	AllowCredentials bool   `mapstructure:"allow_credentials"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWT           JWTConfig           `mapstructure:"jwt"`
	Session       SessionConfig       `mapstructure:"session"`
	Password      PasswordConfig      `mapstructure:"password"`
	Verification  VerificationConfig  `mapstructure:"verification"`
	PasswordReset PasswordResetConfig `mapstructure:"password-reset"`
	CSRF          CSRFConfig          `mapstructure:"csrf"`
	RateLimit     AuthRateLimitConfig `mapstructure:"rate_limit"`
	OAuth         OAuthConfig         `mapstructure:"oauth"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Expiration         string `mapstructure:"expiration"`
	Issuer             string `mapstructure:"issuer"`
	Audience           string `mapstructure:"audience"`
	Algorithm          string `mapstructure:"algorithm"`
	AccessTokenExpiry  string `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry string `mapstructure:"refresh_token_expiry"`
	PrivateKeyPath     string `mapstructure:"private_key_path"`
	PublicKeyPath      string `mapstructure:"public_key_path"`
}

// SessionConfig holds session configuration
type SessionConfig struct {
	MaxActiveSessions int    `mapstructure:"max_active_sessions"`
	ExtendOnActivity  bool   `mapstructure:"extend_on_activity"`
	ActivityThreshold string `mapstructure:"activity_threshold"`
	CleanupInterval   string `mapstructure:"cleanup_interval"`
}

// PasswordConfig holds password policy configuration
type PasswordConfig struct {
	MinLength            int    `mapstructure:"min_length"`
	MaxLength            int    `mapstructure:"max_length"`
	RequireSpecialChars  bool   `mapstructure:"require_special_characters"`
	RequireNumbers       bool   `mapstructure:"require_numbers"`
	RequireUppercase     bool   `mapstructure:"require_uppercase"`
	RequireLowercase     bool   `mapstructure:"require_lowercase"`
	HistoryCount         int    `mapstructure:"history_count"`
	Expiration           string `mapstructure:"expiration"`
	LockoutThreshold     int    `mapstructure:"lockout_threshold"`
	LockoutDuration      string `mapstructure:"lockout_duration"`
	ResetTokenExpiry     string `mapstructure:"reset_token_expiry"`
	ResetTokenLength     int    `mapstructure:"reset_token_length"`
	ResetTokenCharacters string `mapstructure:"reset_token_characters"`
}

// VerificationConfig holds verification configuration
type VerificationConfig struct {
	RequireEmailVerification bool   `mapstructure:"require_email_verification"`
	EmailVerificationExpiry  string `mapstructure:"email_verification_expiry"`
	VerificationTokenLength  int    `mapstructure:"verification_token_length"`
}

// PasswordResetConfig holds password reset configuration
type PasswordResetConfig struct {
	RequirePasswordReset  bool   `mapstructure:"require_password_reset"`
	PasswordResetExpiry   string `mapstructure:"password_reset_expiry"`
	PasswordResetTokenLen int    `mapstructure:"password_reset_token_length"`
}

// CSRFConfig holds CSRF configuration
type CSRFConfig struct {
	TokenLength  int    `mapstructure:"token_length"`
	TokenExpiry  string `mapstructure:"token_expiry"`
	TokenStorage string `mapstructure:"token_storage"`
}

// AuthRateLimitConfig holds authentication rate limit configuration
type AuthRateLimitConfig struct {
	Login         RateLimitRule `mapstructure:"login"`
	Registration  RateLimitRule `mapstructure:"registration"`
	PasswordReset RateLimitRule `mapstructure:"password_reset"`
	Verification  RateLimitRule `mapstructure:"verification"`
	General       RateLimitRule `mapstructure:"general"`
}

// RateLimitRule holds rate limit rule configuration
type RateLimitRule struct {
	Requests int    `mapstructure:"requests"`
	Duration string `mapstructure:"duration"`
}

// OAuthConfig holds OAuth configuration
type OAuthConfig struct {
	Google GoogleOAuthConfig `mapstructure:"google"`
}

// GoogleOAuthConfig holds Google OAuth configuration
type GoogleOAuthConfig struct {
	ClientID    string   `mapstructure:"client_id"`
	RedirectURL string   `mapstructure:"redirect_url"`
	Scopes      []string `mapstructure:"scopes"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level       string            `mapstructure:"level"`
	Format      string            `mapstructure:"format"`
	Output      string            `mapstructure:"output"`
	FilePath    string            `mapstructure:"file_path"`
	Fields      map[string]string `mapstructure:"fields"`
	Sanitize    SanitizeConfig    `mapstructure:"sanitize"`
	HTTP        HTTPLogConfig     `mapstructure:"http"`
	Performance PerformanceConfig `mapstructure:"performance"`
	Stacktrace  StacktraceConfig  `mapstructure:"stacktrace"`
	Rotation    RotationConfig    `mapstructure:"rotation"`
}

// SanitizeConfig holds log sanitization configuration
type SanitizeConfig struct {
	Fields []string `mapstructure:"fields"`
}

// HTTPLogConfig holds HTTP request logging configuration
type HTTPLogConfig struct {
	Enabled   bool          `mapstructure:"enabled"`
	Level     string        `mapstructure:"level"`
	BodyLimit int           `mapstructure:"body_limit"`
	Headers   HeadersConfig `mapstructure:"headers"`
}

// HeadersConfig holds HTTP headers logging configuration
type HeadersConfig struct {
	Include []string `mapstructure:"include"`
	Exclude []string `mapstructure:"exclude"`
}

// PerformanceConfig holds performance logging configuration
type PerformanceConfig struct {
	SamplingRate  float64 `mapstructure:"sampling_rate"`
	SlowThreshold string  `mapstructure:"slow_threshold"`
}

// StacktraceConfig holds stack trace logging configuration
type StacktraceConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Level   string `mapstructure:"level"`
}

// RotationConfig holds log rotation configuration
type RotationConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	MaxSize    int  `mapstructure:"max_size"`
	MaxAge     int  `mapstructure:"max_age"`
	MaxBackups int  `mapstructure:"max_backups"`
	Compress   bool `mapstructure:"compress"`
}

// LoadConfig loads configuration from files and environment variables
func LoadConfig(path string) (*Config, error) {
	var config Config

	viper.AddConfigPath(path)
	viper.SetConfigType("yaml")

	// Load app config
	viper.SetConfigName("app")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading app config: %w", err)
	}

	// Load auth config
	viper.SetConfigName("auth")
	if err := viper.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("error reading auth config: %w", err)
	}

	for _, k := range viper.AllKeys() {
		value := viper.GetString(k)
		if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
			viper.Set(k, _getEnvOrPanic(strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")))
		}
	}

	// Unmarshal config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func _getEnvOrPanic(env string) string {
	res := os.Getenv(env)
	if len(res) == 0 {
		panic("Mandatory env variable not found:" + env)
	}
	return res
}

// GetConfig gets the application configuration
func GetConfig() (*Config, error) {
	config, err := LoadConfig("./configs")
	if err != nil {
		return nil, err
	}
	return config, nil
}
