package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	grafana "github.com/grafana/grafana-api-golang-client"
)

func fetchAlertRuleUID(client *grafana.Client, folderTitle, group, name string) (string, error) {
	var folderUID string
	rules, err := listAlertRules()
	if err != nil {
		return "", fmt.Errorf("获取告警规则失败: %w", err)
	}
	// 先获取folderUID
	folders, err := client.Folders()
	if err != nil {
		return "", fmt.Errorf("获取文件夹列表失败: %w", err)
	}
	// 遍历folders，查找匹配的文件夹
	if len(folders) == 0 {
		return "", fmt.Errorf("没有找到任何文件夹")
	}
	folderUID = ""
	for _, folder := range folders {
		if folder.Title == folderTitle {
			fmt.Printf("已找到文件夹: %s, folder UID: %s\n", folder.Title, folder.UID)
			folderUID = folder.UID // 更新folderTitle为UID
			break
		}
	}
	if folderUID == "" {
		return "", fmt.Errorf("未找到匹配的文件夹: %s", folderTitle)
	}

	for _, rule := range rules {
		fmt.Println(rule.Title, rule.RuleGroup, rule.FolderUID)
		if rule.Title == name && rule.RuleGroup == group {
			return rule.UID, nil
		}
	}
	fmt.Printf("未找到匹配的告警规则: %s, group: %s, folderUID: %s\n", name, group, folderUID)
	return "", fmt.Errorf("未找到匹配的告警规则")
}

func update(client *grafana.Client,
	ruleUID, cpuThreshold, memThreshold, diskThreshold string) (string, error) {
	originalRule, err := client.AlertRule(ruleUID)
	if err != nil {
		log.Fatalf("获取规则失败: %v", err)
	}

	// 3. 修改规则配置（保留UID和Version！）
	updatedRule := originalRule

	// 定义CPU/MEM/DISK
	updatedRule.Title = ALERT_TITLE // 修改标题
	updatedRule.Condition = "A"     // 调整条件引用
	if model, ok := updatedRule.Data[0].Model.(map[string]any); ok {
		model["expr"] = `100 - (avg(irate(node_cpu_seconds_total{mode="idle"}[5m])) by (instance) * 100) > ` + cpuThreshold // 阈值从80%→90%
		updatedRule.Data[0].Model = model
	} else {
		log.Fatalf("Model 字段不是 map[string]interface{} 类型")
	}

	if model, ok := updatedRule.Data[1].Model.(map[string]interface{}); ok {
		model["expr"] = `(1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes) * 100 > ` + memThreshold // 阈值从80%→90%
		updatedRule.Data[1].Model = model
	} else {
		log.Fatalf("Model 字段不是 map[string]interface{} 类型")
	}

	if model, ok := updatedRule.Data[2].Model.(map[string]interface{}); ok {
		model["expr"] = `100 - (node_filesystem_avail_bytes{fstype=~"ext4|xfs",mountpoint="/"} / node_filesystem_size_bytes{fstype=~"ext4|xfs",mountpoint="/"} * 100) > ` + diskThreshold // 阈值从80%→90%
		updatedRule.Data[2].Model = model
	} else {
		log.Fatalf("Model 字段不是 map[string]interface{} 类型")
	}

	// 4. 执行更新
	if err := client.UpdateAlertRule(&updatedRule); err != nil {
		log.Fatalf("更新失败: %v", err)
	}
	fmt.Println("告警规则更新成功！")
	return ruleUID, nil
}

// 创建告警规则
func updateAlertRule(
	client *grafana.Client,
	cpuThreshold, memThreshold, diskThreshold string) (string, error) {
	// 获取数据源UID
	datasource, err := client.DataSourceByName("Prometheus")
	if err != nil {
		return "", fmt.Errorf("获取数据源失败: %w", err)
	}
	fmt.Printf("datasource: %s, UID: %s\n", datasource.Name, datasource.UID)

	ruleUID, _ := fetchAlertRuleUID(client, ALERT_FOLDERTITLE, ALERT_GROUP, ALERT_TITLE)
	fmt.Printf("告警规则UID: %s\n", ruleUID)
	if ruleUID != "" {
		fmt.Println("告警规则已存在，更新规则")
		return update(client, ruleUID, cpuThreshold, memThreshold, diskThreshold)
	} else {
		folder, err := client.NewFolder(ALERT_FOLDERTITLE)
		if err != nil {
			log.Fatalf("创建文件夹失败: %v", err)
		}
		datasource, _ := client.DataSourceByName("Prometheus")

		newRule := grafana.AlertRule{
			ExecErrState: "Error",    // 执行错误状态
			NoDataState:  "OK",       // 无数据状态
			OrgID:        1,          // 组织ID，通常为1
			Updated:      time.Now(), // 更新时间
			Provenance:   "api",      // 来源标识
			IsPaused:     false,      // 是否暂停
			// 其他字段
			Title:     ALERT_TITLE,
			FolderUID: folder.UID,  // 目标文件夹UID
			RuleGroup: ALERT_GROUP, // 规则组名称
			Condition: "C",         // 查询条件标识
			Data: []*grafana.AlertQuery{
				{
					RefID:         "A",
					QueryType:     "",             // 数据源类型
					DatasourceUID: datasource.UID, // Prometheus数据源UID
					RelativeTimeRange: grafana.RelativeTimeRange{
						From: 600,
						To:   0,
					},
					Model: map[string]any{
						"editorMode":   "code",                                                                                      // 编辑器模式
						"refId":        "A",                                                                                         // 引用ID
						"instant":      true,                                                                                        // 是否为瞬时查询
						"editor":       "code",                                                                                      // 编辑器类型
						"expr":         `(1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes) * 100 > ` + memThreshold, // 阈值从80%→90%
						"range":        false,
						"legendFormat": "__auto",
					},
				},
				{
					RefID:         "C",
					QueryType:     "",         // 数据源类型
					DatasourceUID: "__expr__", // 内置表达式数据源
					Model: map[string]any{
						"conditions": []map[string]any{
							{
								"evaluator": map[string]any{
									"type":   "gt", // 大于
									"params": []any{0},
								},
								"operator": map[string]any{
									"type": "and", // 条件连接符
								},
								"query": map[string]any{
									"params": []string{"C"}, // 引用ID
								},
								"reducer": map[string]any{
									"type":   "last",
									"params": []any{},
								},
								"type": "query", // 查询类型
							},
						},
						"datasource": map[string]string{
							"uid":  "__expr__", // 内置表达式数据源
							"type": "__expr__",
						},
						"refId":      "C",         // 引用ID
						"type":       "threshold", // 编辑器类型
						"expression": "A",
					},
				},
			},
			For:         "5m",                                                                        // 持续5分钟触发
			Annotations: map[string]string{"summary": "{{ $labels.instance }} CPU/MEMORY/DISK使用率过高"}, // 告警摘要
			Labels:      map[string]string{"severity": "critical"},                                   // 通知渠道
		}

		//3. bind notification(client)
		err = bindNotification(client)
		if err != nil {
			log.Fatalf("绑定通知失败: %v", err)
		}
		// 4. 创建告警规则
		ruleUID, err := client.NewAlertRule(&newRule)
		if err != nil {
			log.Fatalf("创建失败: %v", err)
		}

		fmt.Printf("✅ 告警规则创建成功！UID: %s\n", ruleUID)
		return ruleUID, nil
	}
}

func listAlertRules() ([]grafana.AlertRule, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
		Timeout: TIMEOUT,
	}
	// 创建 API 请求
	url := fmt.Sprintf("%s%s", GRAFANA_URL, GET_ALERT_RULES_API)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置基本认证
	req.SetBasicAuth(GRAFANA_USER, GRAFANA_PASS)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回错误状态: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result []grafana.AlertRule
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 打印告警规则
	if len(result) == 0 {
		return nil, fmt.Errorf("没有找到任何告警规则")
	}

	return result, nil
}
