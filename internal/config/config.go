package config

  import "flag"

  type Config struct {
      Addr string
  }

  func Parse() Config {
      var cfg Config
      flag.StringVar(&cfg.Addr, "a", "localhost:8080", "адрес сервера")
      flag.Parse()
      return cfg
  }