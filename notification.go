package main

import grafana "github.com/grafana/grafana-api-golang-client"

// type NotificationPolicyTree struct {
//     Receiver       string           `json:"receiver,omitempty"`
//     GroupBy        []string         `json:"group_by,omitempty"`
//     Routes         []SpecificPolicy `json:"routes,omitempty"`
//     GroupWait      string           `json:"group_wait,omitempty"`
//     GroupInterval  string           `json:"group_interval,omitempty"`
//     RepeatInterval string           `json:"repeat_interval,omitempty"`
//     Provenance     string           `json:"provenance,omitempty"`
// }

func bindNotification(client *grafana.Client) error {
	// 创建通知策略
	np := &grafana.NotificationPolicyTree{
		Receiver: "Webhook",
		Routes: []grafana.SpecificPolicy{{
			Receiver: "Webhook",
			ObjectMatchers: grafana.Matchers{
				{Type: grafana.MatchEqual, Name: "severity", Value: "critical"},
			},
		}},
		GroupBy:        []string{"alertname"},
		GroupWait:      "30s",
		GroupInterval:  "5m",
		RepeatInterval: "1h",
	}

	err := client.SetNotificationPolicyTree(np)
	if err != nil {
		return err
	}

	return nil
}
