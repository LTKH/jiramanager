package v1

import (
    "net/http"
    "log"
    "io"
    "fmt"
    "time"
    "net/url"
    "io/ioutil"
    "crypto/sha1"
    "encoding/hex"
    "encoding/json"
    "github.com/ltkh/jiramanager/internal/db"
    "github.com/ltkh/jiramanager/internal/config"
    "github.com/ltkh/jiramanager/internal/template"
    "github.com/andygrunwald/go-jira"
)

var (
    req_chan = make(map[string](chan map[string]interface{}))
)

type Api struct {
    Client       db.DbClient
    Config       *config.Config
    Template     *template.Template
}

type Resp struct {
    Status       string                       `json:"status"`
    Error        string                       `json:"error,omitempty"`
    Warnings     []string                     `json:"warnings,omitempty"`
}

type Data struct {
    Receiver     string                       `json:"receiver"`
    Alerts       []map[string]interface{}     `json:"alerts"`
}

func createIssue(base string, tp jira.BasicAuthTransport, is *jira.Issue) (*jira.Issue, error) {

    jiraClient, err := jira.NewClient(tp.Client(), base)
    if err != nil {
        return is, err
    }

    is, _, err = jiraClient.Issue.Create(is)
    if err != nil {
        return is, err
    }

    return is, nil
}

func getHash(text string) string {
    h := sha1.New()
    io.WriteString(h, text)
    return hex.EncodeToString(h.Sum(nil))
}

func encodeResp(resp *Resp) []byte {
    jsn, err := json.Marshal(resp)
    if err != nil {
        return encodeResp(&Resp{Status:"error", Error:err.Error()})
    }
    return jsn
}

func (api *Api) UpdateStatus() error {
    // Geting issues from database
    issues, err := api.Client.LoadIssues()
    if err != nil {
        return err
    }

    tp := jira.BasicAuthTransport{
        Username: api.Config.Defaults.User,
        Password: api.Config.Defaults.Password,
    }

    for _, i := range issues {

        u, err := url.Parse(i.IssueSelf)
        if err != nil {
            log.Printf("[error] %v", err)
            continue
        }

        base := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
        jiraClient, err := jira.NewClient(tp.Client(), base)
        if err != nil {
            log.Printf("[error] %v", err)
            continue
        }
        
        issue, _, err := jiraClient.Issue.Get(i.IssueKey, nil)
        if err != nil {
            log.Printf("[error] %v", err)
            continue
        }

        if i.StatusId != issue.Fields.Status.ID {
            if err := api.Client.UpdateStatus(i.GroupId, issue.Fields.Status.ID, issue.Fields.Status.Name); err != nil {
                log.Printf("[error] %v", err)
                continue
            }
            log.Printf("[info] issue status updated: %s", i.IssueKey)
        }

        if i.Updated + 600 < time.Now().UTC().Unix() {
            for _, s := range api.Config.Defaults.ResolveState {
                if i.StatusId == s {
                    if err := api.Client.DeleteIssue(i.GroupId); err != nil {
                        log.Printf("[error] %v", err)
                        continue
                    }
                    log.Printf("[info] issue removed from database: %s", i.IssueKey)
                }
            }
        }

    }

    return nil
}

func (api *Api) ApiAlerts(w http.ResponseWriter, r *http.Request) {
    var data Data

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Printf("[error] %v - %s", err, r.URL.Path)
        w.WriteHeader(400)
        w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
        return
    }

    if err := json.Unmarshal(body, &data); err != nil {
        log.Printf("[error] %v - %s", err, r.URL.Path)
        w.WriteHeader(400)
        w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
        return
    }

    for _, alert := range data.Alerts {
        go func(alert map[string]interface{}){
            req_chan[data.Receiver] <- alert 
        }(alert)
    }
    
    w.Write(encodeResp(&Resp{Status:"success"}))
    return
}

func (api *Api) readChannel(receiver *config.Receiver) {
    for {

        for i := 0; i < len(req_chan[receiver.Name]); i++ {
            select {
                case alert := <- req_chan[receiver.Name]:

                    labels, err := json.Marshal(alert["labels"])
                    if err != nil {
                        log.Printf("[error] read alert %v", err)
                        continue
                    }

                    if alert["status"] == "resolved" {
                        continue
                    }

                    group_id := getHash(string(labels))

                    task, err := api.Client.LoadIssue(group_id)
                    if err != nil {
                        log.Printf("[error] %v", err)
                        continue
                    }

                    if task.GroupId != "" {
                        continue
                    }

                    issueSummary, err := api.Template.Execute(receiver.Summary, alert)
                    if err != nil {
                        log.Printf("[error] %v", err)
                        continue
                    }

                    issueDesc, err := api.Template.Execute(receiver.Description, alert)
                    if err != nil {
                        log.Printf("[error] %v", err)
                        continue
                    }

                    is := &jira.Issue{
                        Fields: &jira.IssueFields{
                            Project:     receiver.Project,
                            Type:        receiver.IssueType,
                            Summary:     issueSummary,
                            Description: issueDesc,
                        },
                    }

                    if len(receiver.Components) > 0 {
                        for _, component := range receiver.Components {
                            is.Fields.Components = append(is.Fields.Components, &component)
                        }
                    }

                    if len(receiver.Fields) > 0 {
                        //for key, value := range receiver.Fields {
                        //    is.Fields.Unknowns[key] = value
                        //}
                    }

                    tp := jira.BasicAuthTransport{
                        Username: api.Config.Defaults.User,
                        Password: api.Config.Defaults.Password,
                    }

                    is, err = createIssue(receiver.ApiUrl, tp, is)
                    if err != nil {
                        log.Printf("[error] create issue %v", err)
                        continue
                    }

                    tk := config.Issue{
                        GroupId:    group_id,
                        IssueId:    is.ID,
                        IssueKey:   is.Key,
                        IssueSelf:  is.Self,
                    }
                    if err := api.Client.SaveIssue(tk); err != nil {
                        log.Printf("[error] save issue %v", err)
                        continue
                    }

                    log.Printf("[info] create issue: %s", is.Key)
                    

                default:
                    continue
            }
        }

        time.Sleep(5 * time.Second)
    }
}

func New(config *config.Config) (*Api, error) {
    // Connection to data base
    client, err := db.NewClient(config.DB)
    if err != nil {
        return nil, err
    }

    tmpl, err := template.LoadTemplate(config.Template)
    if err != nil {
        return nil, err
    }

    api := &Api{ Client: client, Config: config, Template: tmpl }
    for _, receiver := range api.Config.Receivers {
        req_chan[receiver.Name] = make(chan map[string]interface{}, 1000)
        go api.readChannel(receiver)
    }
    
    return api, nil
}