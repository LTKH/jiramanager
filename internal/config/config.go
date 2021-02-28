package config

import (
	"fmt"
	"net/url"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/andygrunwald/go-jira"
)

type Config struct {
	Defaults         *Receiver               `yaml:"defaults"`
	DB               *DB                     `yaml:"db"`
	Receivers        []*Receiver             `yaml:"receivers"`
	Template         string                  `yaml:"template"`
}

type DB struct {
	Client           string                  `yaml:"client"`
	ConnString       string                  `yaml:"conn_string"`
	IssuesTable      string                  `yaml:"issues_table"`
}

type Receiver struct {
	Name             string                  `yaml:"name"`
	ApiUrl           string                  `yaml:"api_url"`
	User             string                  `yaml:"user"`
	Password         string                  `yaml:"password"`
	Project          jira.Project            `yaml:"project"`
	IssueType        jira.IssueType          `yaml:"issue_type"`
	Priority         jira.Priority           `yaml:"priority"`
	Summary          string                  `yaml:"summary"`
	Description      string                  `yaml:"description"`
	Components       []jira.Component        `yaml:"components"`
	Fields           map[string]interface{}  `yaml:"fields"`
}

type Issue struct {
	GroupId          string
	StatusId         string
	StatusName       string
	IssueId          string
	IssueKey         string
	IssueSelf        string
	Created          int64
	Updated          int64
	Template         string
}

func New(filename string) (*Config, error) {
	cfg := &Config{}

    content, err := ioutil.ReadFile(filename)
    if err != nil {
       return cfg, err
    }

    if err := yaml.UnmarshalStrict(content, cfg); err != nil {
        return cfg, err
	}

	for _, rc := range cfg.Receivers {
        if rc.Name == "" {
			return cfg, fmt.Errorf("missing name for receiver %+v", rc)
		}

		// Check API access fields
        if rc.ApiUrl == "" {
			if cfg.Defaults.ApiUrl == "" {
				return cfg, fmt.Errorf("missing api_url in receiver %q", rc.Name)
			}
			rc.ApiUrl = cfg.Defaults.ApiUrl
		}
		if _, err := url.Parse(rc.ApiUrl); err != nil {
			return cfg, fmt.Errorf("invalid api_url %q in receiver %q: %s", rc.ApiUrl, rc.Name, err)
		}
		if rc.User == "" {
			if cfg.Defaults.User == "" {
				return cfg, fmt.Errorf("missing user in receiver %q", rc.Name)
			}
			rc.User = cfg.Defaults.User
		}
		if rc.Password == "" {
			if cfg.Defaults.Password == "" {
				return cfg, fmt.Errorf("missing password in receiver %q", rc.Name)
			}
			rc.Password = cfg.Defaults.Password
		}

		// Check required issue fields
		if rc.Project.ID == "" && rc.Project.Key == "" && rc.Project.Name == "" {
			if cfg.Defaults.Project.ID == "" && cfg.Defaults.Project.Key == "" && cfg.Defaults.Project.Name == "" {
				return cfg, fmt.Errorf("missing project in receiver %q", rc.Name)
			}
			rc.Project = cfg.Defaults.Project
		}
		if rc.IssueType.ID == "" && rc.IssueType.Name == "" {
			if cfg.Defaults.IssueType.ID == "" && cfg.Defaults.IssueType.Name == "" {
				return cfg, fmt.Errorf("missing issue_type in receiver %q", rc.Name)
			}
			rc.IssueType = cfg.Defaults.IssueType
		}
		if rc.Summary == "" {
			if cfg.Defaults.Summary == "" {
				return cfg, fmt.Errorf("missing summary in receiver %q", rc.Name)
			}
			rc.Summary = cfg.Defaults.Summary
		}

		// Populate optional issue fields, where necessary
		if rc.Priority.ID == "" && rc.Priority.Name == "" {
			if cfg.Defaults.Priority.ID != "" || cfg.Defaults.Priority.Name != "" {
				rc.Priority = cfg.Defaults.Priority
			}
		}
		if rc.Description == "" && cfg.Defaults.Description != "" {
			rc.Description = cfg.Defaults.Description
		}
		if len(cfg.Defaults.Fields) > 0 {
			for key, value := range cfg.Defaults.Fields {
				if _, ok := rc.Fields[key]; !ok {
					rc.Fields[key] = value
				}
			}
		}
	}
	
	return cfg, nil
}

