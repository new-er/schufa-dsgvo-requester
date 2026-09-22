package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	_ "embed"
)

//go:embed config.toml.example
var exampleConfig string

type MailConfig struct {
	SMTPHost string `toml:"smtp_host"`
	SMTPPort int    `toml:"smtp_port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	FromAddr string `toml:"from_addr"`
	ToAddr   string `toml:"to_addr"`
}

type PersonalConfig struct {
	Vorname      string `toml:"vorname"`
	Nachname     string `toml:"nachname"`
	Geburtsdatum string `toml:"geburtsdatum"`
	Strasse      string `toml:"strasse"`
	PLZ          string `toml:"plz"`
	Wohnort      string `toml:"wohnort"`
	Land         string `toml:"land"`
}

type ScheduleConfig struct {
	IntervalMonths int    `toml:"interval_months"`
	LastSent       string `toml:"last_sent"`
}

type Config struct {
	Mail      MailConfig      `toml:"mail"`
	Personal  PersonalConfig  `toml:"personal"`
	Schedule  ScheduleConfig  `toml:"schedule"`
}

func Load(path string) (*Config, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fi.Mode().Perm()&077 != 0 {
		return nil, &ValidationError{"config file permissions too open (must be 0600): " + path}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, err
	}
	// defaults
	if cfg.Schedule.IntervalMonths == 0 {
		cfg.Schedule.IntervalMonths = 3
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Mail.SMTPHost == "" || c.Mail.SMTPPort == 0 || c.Mail.Username == "" || c.Mail.Password == "" || c.Mail.FromAddr == "" || c.Mail.ToAddr == "" {
		return &ValidationError{"mail section incomplete"}
	}
	if c.Personal.Vorname == "" || c.Personal.Nachname == "" || c.Personal.Geburtsdatum == "" || c.Personal.Strasse == "" || c.Personal.PLZ == "" || c.Personal.Wohnort == "" || c.Personal.Land == "" {
		return &ValidationError{"personal section incomplete"}
	}
	return nil
}

type ValidationError struct {
	msg string
}

func (e *ValidationError) Error() string {
	return e.msg
}

func Save(path string, cfg *Config) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := toml.NewEncoder(f)
	return enc.Encode(cfg)
}

func LoadOrCreate(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte(exampleConfig), 0600); err != nil {
			return nil, err
		}
		fmt.Printf("Created %s from example — please edit it\n", path)
	}
	return Load(path)
}