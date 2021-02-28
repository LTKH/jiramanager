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
	conn, err := sql.Open("mysql", conf.ConnString)
	if err != nil {
		return nil, err
	}
	return &Client{ client: conn, config: conf }, nil
}

func (db *Client) LoadIssue(group_id string) (config.Issue, error) {
    var issue config.Issue

    stmt, err := db.client.Prepare(fmt.Sprintf(
		"select group_id,status_id,status_name,issue_id,issue_key,issue_self from %s where group_id = ?", 
		db.config.IssuesTable,
	))
	if err != nil {
		return issue, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(group_id).Scan(&issue.GroupId, &issue.StatusId, &issue.StatusName, &issue.IssueId, &issue.IssueKey, &issue.IssueSelf)
	if err != nil {
		return issue, nil
	}

  	return issue, nil
}

func (db *Client) LoadIssues() ([]config.Issue, error) {
	var result []config.Issue

	rows, err := db.client.Query(fmt.Sprintf(
		"select group_id,status_id,status_name,issue_id,issue_key,issue_self,created,updated from %s", 
		db.config.IssuesTable,
	))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var issue config.Issue
        err := rows.Scan(&issue.GroupId, &issue.StatusId, &issue.StatusName, &issue.IssueId, &issue.IssueKey, &issue.IssueSelf, &issue.Created, &issue.Updated)
        if err != nil {
            return nil, err
		}
		result = append(result, issue) 
    }

  	return result, nil
}

func (db *Client) SaveIssue(issue config.Issue) error {
	stmt, err := db.client.Prepare(fmt.Sprintf(
		"replace into %s (group_id,status_id,status_name,issue_id,issue_key,issue_self,created,updated,template) values (?,?,?,?,?,?,?,?,?)", 
		db.config.IssuesTable,
	))
	if err != nil {
		return err
	}
	defer stmt.Close()

	utc := time.Now().UTC().Unix()
	_, err = stmt.Exec(issue.GroupId, issue.StatusId, issue.StatusName, issue.IssueId, issue.IssueKey, issue.IssueSelf, utc, utc, issue.Template)
	if err != nil {
		return err
	}

	return nil
}

func (db *Client) UpdateStatus(group_id, status_id, status_name string) error {
	stmt, err := db.client.Prepare(fmt.Sprintf(
		"update %s set status_id = ?, status_name = ?, updated = ? where group_id = ?", 
		db.config.IssuesTable,
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

func (db *Client) DeleteIssue(group_id string) error {

	stmt, err := db.client.Prepare(fmt.Sprintf(
		"delete from %s where group_id = ?", 
		db.config.IssuesTable,
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