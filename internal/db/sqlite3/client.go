package sqlite3

import (
    "time"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "github.com/ltkh/jiramanager/internal/config"
)

type Client struct {
    client *sql.DB
    config *config.DB
}

func NewClient(conf *config.DB) (*Client, error) {
    conn, err := sql.Open("sqlite3", conf.ConnString)
    if err != nil {
        return nil, err
    }
    return &Client{ client: conn, config: conf }, nil
}

func (db *Client) Close() error {
	db.client.Close()

	return nil
}

func (db *Client) CreateTables() error {
    _, err := db.client.Exec(
	  `create table if not exists issues (
		group_id      varchar(50) primary key,
		status_id     varchar(50) default '',
		status_name   varchar(100) default '',
		issue_id      varchar(50),
		issue_key     varchar(50),
		issue_self    varchar(1500),
		created       bigint(20) default 0,
		updated       bigint(20) default 0,
		template      varchar(250)
	  );`)
    if err != nil {
        return err
    }

    return nil
}

func (db *Client) LoadIssue(group_id string) (config.Issue, error) {
    var issue config.Issue

    stmt, err := db.client.Prepare("select group_id,status_id,status_name,issue_id,issue_key,issue_self from issues where group_id = ?")
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

    rows, err := db.client.Query("select group_id,status_id,status_name,issue_id,issue_key,issue_self,created,updated from issues")
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
    stmt, err := db.client.Prepare("replace into issues (group_id,status_id,status_name,issue_id,issue_key,issue_self,created,updated,template) values (?,?,?,?,?,?,?,?,?)")
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
    stmt, err := db.client.Prepare("update issues set status_id = ?, status_name = ?, updated = ? where group_id = ?")
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

    stmt, err := db.client.Prepare("delete from issues where group_id = ?")
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
