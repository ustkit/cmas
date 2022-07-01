package config

type Config struct {
	Sever          string `env:"ADDRESS"`
	PollInterval   string `env:"POLL_INTERVAL"`
	ReportInterval string `env:"REPORT_INTERVAL"`
	DataType       string
	Key            string `env:"KEY"`
}
