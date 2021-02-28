package db

import (
	"errors"
	"github.com/ltkh/jiramanager/internal/config"
	"github.com/ltkh/jiramanager/internal/db/mysql"
)

type DbClient interface {
	LoadIssue(mgrp_id string) (config.Issue, error)
	LoadIssues() ([]config.Issue, error)
	SaveIssue(issue config.Issue) error
	UpdateStatus(group_id, status_id, status_name string) error
	DeleteIssue(group_id string) error
}

func NewClient(config *config.DB) (DbClient, error) {
	switch config.Client {
	    case "mysql":
            return mysql.NewClient(config)
	}
	return nil, errors.New("invalid client")
}