package main

import "time"

const (
	GRAFANA_URL         = "http://10.67.100.103:32265"
	GRAFANA_USER        = "admin"
	GRAFANA_PASS        = "admin"
	WEBHOOK_URL         = "http://10.67.100.103:8080"
	TIMEOUT             = 15 * time.Second // 请求超时时间
	GET_ALERT_RULES_API = "/api/v1/provisioning/alert-rules"
	ALERT_FOLDERTITLE   = "joiningos"
	ALERT_GROUP         = "group"
	ALERT_TITLE         = "高CPU/MEMORY/DISK使用率告警（新阈值）"
)
