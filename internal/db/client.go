package db

import (
    "errors"
    "github.com/ltkh/jiramanager/internal/config"
    "github.com/ltkh/jiramanager/internal/db/mysql"
    "github.com/ltkh/jiramanager/internal/db/sqlite3"
)

type DbClient interface {
    CreateTables() error
    LoadIssue(mgrp_id string) (config.Issue, error)
    LoadIssues() ([]config.Issue, error)
    SaveIssue(issue config.Issue) error
    UpdateStatus(group_id, status_id, status_name string) error
    DeleteIssue(group_id string) error
    Close() error
}

func NewClient(config *config.DB) (DbClient, error) {
    switch config.Client {
        case "mysql":
            return mysql.NewClient(config)
        case "sqlite3":
            return sqlite3.NewClient(config)
    }
    return nil, errors.New("invalid client")
}