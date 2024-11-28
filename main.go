package main

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Settings for reading from YAML
type Config struct {
	RedmineURLs []RedmineConfig `yaml:"redmine_urls"`
}

type RedmineConfig struct {
	URL     string `yaml:"url"`
	APIKey  string `yaml:"api_key"`
	QueryID int    `yaml:"query_id"`
	Limit   int    `yaml:"limit"`
}

// Response from Redmine API
type RedmineResponse struct {
	Issues []struct {
		ID         int    `json:"id"`
		Project    Named  `json:"project"`
		Tracker    Named  `json:"tracker"`
		Status     Named  `json:"status"`
		AssignedTo Named  `json:"assigned_to"`
		Subject    string `json:"subject"`
	} `json:"issues"`
}

type Named struct {
	Name string `json:"name"`
}

// Ticket information for display
type Issue struct {
	Ticket  int
	Project string
	Tracker string
	Status  string
	Name    string
	Subject string
	Link    string
}

// Maintains a list of tickets for each person in charge
type UserIssues map[string][]Issue

func loadConfig(filename string) (Config, error) {
	var config Config
	file, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(file, &config)
	return config, err
}

func fetchRedmineIssues(config RedmineConfig) (UserIssues, error) {
	issues := make(UserIssues)
	client := &http.Client{Timeout: 5 * time.Second}

	url := config.URL + "issues.json" +
		"?key=" + config.APIKey +
		"&query_id=" + string(config.QueryID) +
		"&limit=" + string(config.Limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var redmineResp RedmineResponse
	if err := json.Unmarshal(body, &redmineResp); err != nil {
		return nil, err
	}

	for _, item := range redmineResp.Issues {
		issue := Issue{
			Ticket:  item.ID,
			Project: item.Project.Name,
			Tracker: item.Tracker.Name,
			Status:  item.Status.Name,
			Name:    item.AssignedTo.Name,
			Subject: item.Subject,
			Link:    config.URL,
		}
		issues[item.AssignedTo.Name] = append(issues[item.AssignedTo.Name], issue)
	}

	return issues, nil
}

func combineRedmineIssues(configs []RedmineConfig) (UserIssues, error) {
	combined := make(UserIssues)

	for _, config := range configs {
		issues, err := fetchRedmineIssues(config)
		if err != nil {
			return nil, err
		}

		for name, userIssues := range issues {
			combined[name] = append(combined[name], userIssues...)
		}
	}

	return combined, nil
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>今日の活動</title>
    <link rel="icon" href="data:,">
    <meta charset="utf-8">
</head>
<body>
    <table border="1">
    <tr>
    <th>チケット</th>
    <th>プロジェクト</th>
    <th>トラッカー</th>
    <th>ステータス</th>
    <th>題名</th>
    <th>担当者</th>
    </tr>
    {{range $name, $issues := .}}
        {{range $issue := $issues}}
        <tr>
            <td><a href="{{$issue.Link}}issues/{{$issue.Ticket}}">{{$issue.Ticket}}</a></td>
            <td>{{$issue.Project}}</td>
            <td>{{$issue.Tracker}}</td>
            <td>{{$issue.Status}}</td>
            <td>{{$issue.Subject}}</td>
            <td>{{$name}}</td>
        </tr>
        {{end}}
    {{end}}
    </table>
</body>
</html>
`

func handleRoot(tmpl *template.Template, config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		issues, err := combineRedmineIssues(config.RedmineURLs)
		if err != nil {
			http.Error(w, "Failed to fetch Redmine issues: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, issues); err != nil {
			log.Printf("Template execution error: %v", err)
		}
	}
}

func main() {
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	tmpl, err := template.New("issues").Parse(htmlTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	http.HandleFunc("/", handleRoot(tmpl, config))

	port := ":8000"
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
