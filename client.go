package main

import (
	"net/http"
	"net/url"
	"time"

	grafana "github.com/grafana/grafana-api-golang-client"
)

// AlertRule represents a Grafana Alert Rule.
type NewAlertRule struct {
	Annotations  map[string]string     `json:"annotations,omitempty"`
	Condition    string                `json:"condition"`
	Data         []*grafana.AlertQuery `json:"data"`
	ExecErrState grafana.ExecErrState  `json:"execErrState"`
	FolderUID    string                `json:"folderUid"`
	ID           int64                 `json:"id,omitempty"`
	Labels       map[string]string     `json:"labels,omitempty"`
	NoDataState  grafana.NoDataState   `json:"noDataState"`
	OrgID        int64                 `json:"orgId"`
	RuleGroup    string                `json:"ruleGroup"`
	Title        string                `json:"title"`
	UID          string                `json:"uid,omitempty"`
	Updated      time.Time             `json:"updated"`
	For          string                `json:"for"`
	ForDuration  time.Duration         `json:"-"`
	Provenance   string                `json:"provenance"`
	IsPaused     bool                  `json:"isPaused"`
}

func createGrafanaClient() (*grafana.Client, error) {
	cfg := grafana.Config{
		BasicAuth: url.UserPassword(GRAFANA_USER, GRAFANA_PASS),
		Client:    http.DefaultClient,
	}

	return grafana.New(GRAFANA_URL, cfg)
}
