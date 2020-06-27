package config

import (
	"time"
	"os"
	"github.com/naoina/toml"
)

type Config struct {
	DB                   DB
	Server struct {
		Conf_dir         string
		Log_max_size     int
		Log_max_backups  int
		Log_max_age      int
		Log_compress     bool
	}
	Alerts struct {
		Interval         time.Duration
		Login            string
		Passwd           string
	}
	
	Jira struct {
		Jira_api         string
		Login            string
		Passwd           string
	}
	Monit struct {
		Listen           string
	}
}

type DB struct {
	Client               string
	Conn_string          string
	Tasks_table          string
}

type Task struct {
	Group_id             string
	Task_id              string
	Task_key             string
    Task_self            string
}

func New(filename string) (cfg Config, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	return cfg, toml.NewDecoder(f).Decode(&cfg)
}
