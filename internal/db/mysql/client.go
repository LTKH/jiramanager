package mysql

import (
	"fmt"
	"time"
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
		"select group_id,status_id,status_name,task_id,task_key,task_self from %s where group_id = ?", 
		db.config.Tasks_table,
	))
	if err != nil {
		return task, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(group_id).Scan(&task.Group_id, &task.Status_id, &task.Status_name, &task.Task_id, &task.Task_key, &task.Task_self)
	if err != nil {
		return task, nil
	}

  	return task, nil
}

func (db *Client) LoadTasks() ([]config.Task, error) {
	var result []config.Task

	rows, err := db.client.Query(fmt.Sprintf(
		"select group_id,status_id,status_name,task_id,task_key,task_self,created,updated from %s", 
		db.config.Tasks_table,
	))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task config.Task
        err := rows.Scan(&task.Group_id, &task.Status_id, &task.Status_name, &task.Task_id, &task.Task_key, &task.Task_self, &task.Created, &task.Updated)
        if err != nil {
            return nil, err
		}
		result = append(result, task) 
    }

  	return result, nil
}

func (db *Client) SaveTask(task config.Task) error {
	stmt, err := db.client.Prepare(fmt.Sprintf(
		"replace into %s (group_id,status_id,status_name,task_id,task_key,task_self,created,updated,template) values (?,?,?,?,?,?,?,?,?)", 
		db.config.Tasks_table,
	))
	if err != nil {
		return err
	}
	defer stmt.Close()

	utc := time.Now().UTC().Unix()
	_, err = stmt.Exec(task.Group_id, task.Status_id, task.Status_name, task.Task_id, task.Task_key, task.Task_self, utc, utc, task.Template)
	if err != nil {
		return err
	}

	return nil

}

func (db *Client) UpdateStatus(group_id, status_id, status_name string) error {
	stmt, err := db.client.Prepare(fmt.Sprintf(
		"update %s set status_id = ?, status_name = ?, updated = ? where group_id = ?", 
		db.config.Tasks_table,
	))
	if err != nil {
		return err
	}
	defer stmt.Close()

	utc := time.Now().UTC().Unix()
	_, err = stmt.Exec(status_id, status_name, utc, group_id)
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