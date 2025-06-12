package main

import "time"

const (
	GRAFANA_URL         = "http://grafana.default:80"
	GRAFANA_USER        = "admin"
	GRAFANA_PASS        = "admin"
	TIMEOUT             = 15 * time.Second // 请求超时时间
	GET_ALERT_RULES_API = "/api/v1/provisioning/alert-rules"
	ALERT_FOLDERTITLE   = "joiningos"
	ALERT_GROUP         = "group"
)
