package config

// Config содержит настройки сервера.
type Config struct {
	Address       string `env:"ADDRESS"`
	StoreInterval string `env:"STORE_INTERVAL"`
	StoreFile     string `env:"STORE_FILE"`
	Restore       bool   `env:"RESTORE"`
	Key           string `env:"KEY"`
	DataBaseDSN   string `env:"DATABASE_DSN"`
}
