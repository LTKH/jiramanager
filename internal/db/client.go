package db

import (
	"errors"
	"github.com/ltkh/jiramanager/internal/config"
	"github.com/ltkh/jiramanager/internal/db/mysql"
)

type DbClient interface {
	LoadTask(mgrp_id string) (config.Task, error)
	LoadTasks() ([]config.Task, error)
	SaveTask(task config.Task) error
	UpdateStatus(group_id, status_id, status_name string) error
	DeleteTask(group_id string) error
}

func NewClient(config *config.DB) (DbClient, error) {
	switch config.Client {
	    case "mysql":
            return mysql.NewClient(config)
	}
	return nil, errors.New("invalid client")
}