package mysql

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ltkh/jiramanager/internal/config"
)

type Client struct {
	client   *sql.DB
    config   *config.DB
}

func NewClient(conf *config.DB) (*Client, error) {
	conn, err := sql.Open("mysql", conf.Conn_string)
	if err != nil {
		return nil, err
	}
	return &Client{ client: conn, config: conf }, nil
}

func (db *Client) LoadTask(group_id string) (config.Task, error) {
    var task config.Task

    stmt, err := db.client.Prepare(fmt.Sprintf(
		"select group_id,task_id,task_key,task_self from %s where group_id = ?", 
		db.config.Tasks_table,
	))
	if err != nil {
		return task, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(group_id).Scan(&task.Group_id, &task.Task_id, &task.Task_key, &task.Task_self)
	if err != nil {
		return task, nil
	}

  	return task, nil
}

func (db *Client) LoadTasks() ([]config.Task, error) {
	var result []config.Task

	rows, err := db.client.Query(fmt.Sprintf(
		"select group_id,task_id,task_key,task_self from %s", 
		db.config.Tasks_table,
	))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task config.Task
        err := rows.Scan(&task.Group_id, &task.Task_id, &task.Task_key, &task.Task_self)
        if err != nil {
            return nil, err
		}
		result = append(result, task) 
    }

  	return result, nil
}

func (db *Client) SaveTask(task config.Task) error {
	stmt, err := db.client.Prepare(fmt.Sprintf(
		"replace into %s (group_id,task_id,task_key,task_self) values (?,?,?,?)", 
		db.config.Tasks_table,
	))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(task.Group_id, task.Task_id, task.Task_key, task.Task_self)
	if err != nil {
		return err
	}

	return nil

}

func (db *Client) DeleteTask(group_id string) error {

	stmt, err := db.client.Prepare(fmt.Sprintf(
		"delete from %s where group_id = ?", 
		db.config.Tasks_table,
	))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(group_id)
	if err != nil {
		return err
	}

	return nil
}