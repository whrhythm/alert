package main

import (
	"os"

	grafana "github.com/grafana/grafana-api-golang-client"
)

// ptrString returns a pointer to the given string.
func ptrString(s string) *string {
	return &s
}

func ensureWebhookContactPoint(client *grafana.Client) (string, error) {
	// 检查是否已存在Webhook通知渠道
	contactPoint, err := client.ContactPoint("webhook")
	if err == nil {
		return contactPoint.UID, nil
	}

	// 如果不存在，则创建新的Webhook通知渠道
	uid, err := createWebhookContactPoint(client)
	if err != nil {
		return "", err
	}

	return uid, nil
}

// 创建Webhook通知渠道
func createWebhookContactPoint(client *grafana.Client) (string, error) {
	// 配置通知消息模板
	// messageTemplate := `{
	// 	"hostname": "{{ .Labels.kubernetes_io_hostname }}",
	// 	"type": ptrString("webhook"),
	// }`

	// 创建通知渠道配置
	WEBHOOK_URL := os.Getenv("WEBHOOK")
	contactPoint := grafana.ContactPoint{
		Name:     "Webhook",
		Type:     "webhook",
		UID:      "webhook",
		Settings: map[string]any{"url": WEBHOOK_URL},
	}

	// 创建 contact_points 客户端
	return client.NewContactPoint(&contactPoint)
}
