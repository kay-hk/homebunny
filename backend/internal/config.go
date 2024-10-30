package internal

// AppConfig holds configuration for the entire application.
type AppConfig struct {
	RabbitMQ struct {
		User     string `yaml:"User"`
		Password string `yaml:"Password"`
		Host     string `yaml:"Host"`
		VHost    string `yaml:"VHost"`
	} `yaml:"RabbitMQ"`

	Database struct {
		Host     string `yaml:"Host"`
		Port     string `yaml:"Port"`
		User     string `yaml:"User"`
		Password string `yaml:"Password"`
		DBName   string `yaml:"DBName"`
	} `yaml:"Database"`

	Server struct {
		Port string `yaml:"Port"`
	} `yaml:"Server"`

	Producer struct {
		Queue string `yaml:"Queue"`
	} `yaml:"Producer"`

	Consumer struct {
		Queue        string `yaml:"Queue"`
		PrefetchCount int   `yaml:"PrefetchCount"`
	} `yaml:"Consumer"`
}