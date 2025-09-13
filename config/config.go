package config

type WebRTCSignalingAppConfig struct {
	ServerAddr string `env:"SERVER_ADDR" envDefault:"0.0.0.0:8080"`
	LogLevel   string `env:"LOG_LEVEL" envDefault:"info"`
}
