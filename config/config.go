package config

type WebRTCSignalingAppConfig struct {
	ServerAddr     string   `env:"SERVER_ADDR" envDefault:":8080"`
	AllowedOrigins []string `env:"ALLOWED_ORIGINS" envDefault:"*"`
	LogLevel       string   `env:"LOG_LEVEL" envDefault:"info"`
	TURNKey        string   `env:"TURN_KEY" envDefault:""`
	TURNAPIToken   string   `env:"TURN_API_TOKEN" envDefault:""`
}
