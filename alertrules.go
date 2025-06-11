package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	grafana "github.com/grafana/grafana-api-golang-client"
)

func fetchAlertRuleUID(client *grafana.Client, id int64) (grafana.AlertRule, error) {
	rules, err := listAlertRules()
	if err != nil {
		return grafana.AlertRule{}, fmt.Errorf("获取告警规则失败: %w", err)
	}

	for _, rule := range rules {
		if rule.ID == id {
			return rule, nil
		}
	}
	fmt.Printf("未找到匹配的告警ID: %d\n", id)
	return grafana.AlertRule{}, fmt.Errorf("未找到匹配的告警规则")
}

func update(client *grafana.Client, rule grafana.AlertRule, r *AlertRequest) (string, error) {

	code := strings.ToLower(r.MetricCode)
	switch code {
	case "cpu":
		rule.Data[0].Model.(map[string]any)["expr"] = `100 - (avg(irate(node_cpu_seconds_total{mode="idle"}[5m])) by (instance) * 100) ` + r.Operator + r.MetricThreshold // 阈值从80%→90%
	case "memory":
		rule.Data[0].Model.(map[string]any)["expr"] = `(1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes) * 100 ` + r.Operator + r.MetricThreshold // 阈值从80%→90%
	case "disk":
		rule.Data[0].Model.(map[string]any)["expr"] = `100 - (node_filesystem_avail_bytes{fstype=~"ext4|xfs",mountpoint="/"} / node_filesystem_size_bytes{fstype=~"ext4|xfs",mountpoint="/"} * 100) ` + r.Operator + r.MetricThreshold // 阈值从80%→90%
	default:
		return "", fmt.Errorf("不支持的 MetricCode: %s", r.MetricCode)
	}

	// 执行更新
	if err := client.UpdateAlertRule(&rule); err != nil {
		return "", fmt.Errorf("更新告警规则失败: %w", err)
	}
	fmt.Println("告警规则更新成功！")
	return rule.UID, nil
}

func newAlertRule(folder grafana.Folder, datasource grafana.DataSource, r AlertRequest) *grafana.AlertRule {
	var expr string

	code := strings.ToLower(r.MetricCode)
	id, err := strconv.ParseInt(r.RuleID, 10, 64)
	if err != nil {
		return nil
	}

	switch code {
	case "cpu":
		expr = `100 - (avg(irate(node_cpu_seconds_total{mode="idle"}[5m])) by (instance) * 100) ` + r.Operator + r.MetricThreshold
	case "memory":
		expr = `(1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes) * 100 ` + r.Operator + r.MetricThreshold
	case "disk":
		expr = `100 - (node_filesystem_avail_bytes{fstype=~"ext4|xfs",mountpoint="/"} / node_filesystem_size_bytes{fstype=~"ext4|xfs",mountpoint="/"} * 100) ` + r.Operator + r.MetricThreshold
	default:
		return nil
	}

	return &grafana.AlertRule{
		ID:           id,
		ExecErrState: "Error",    // 执行错误状态
		NoDataState:  "OK",       // 无数据状态
		OrgID:        1,          // 组织ID，通常为1
		Updated:      time.Now(), // 更新时间
		Provenance:   "api",      // 来源标识
		IsPaused:     false,      // 是否暂停
		// 其他字段
		Title:     fmt.Sprintf("高 %s 使用率告警", r.MetricCode), // 告警标题
		FolderUID: folder.UID,                              // 目标文件夹UID
		RuleGroup: ALERT_GROUP,                             // 规则组名称
		Condition: "C",                                     // 查询条件标识
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
					"editorMode":   "code", // 编辑器模式
					"refId":        "A",    // 引用ID
					"instant":      true,   // 是否为瞬时查询
					"editor":       "code", // 编辑器类型
					"expr":         expr,   // PromQL表达式
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
		For:         "5m", // 持续5分钟触发
		Annotations: map[string]string{"summary": "{{ $labels.instance }}" + r.MetricCode + "使用率过高"},
		Labels:      map[string]string{"severity": "critical"}, // 通知渠道
	}
}

// 创建告警规则
func updateAlertRule(
	client *grafana.Client, r *AlertRequest) (string, error) {
	id, err := strconv.ParseInt(r.RuleID, 10, 64)
	if err != nil {
		log.Fatalf("RuleID 解析失败: %v", err)
	}
	// 获取数据源UID
	datasource, err := client.DataSourceByName("Prometheus")
	if err != nil {
		return "", fmt.Errorf("获取数据源失败: %w", err)
	}
	fmt.Printf("datasource: %s, UID: %s\n", datasource.Name, datasource.UID)

	rule, err := fetchAlertRuleUID(client, id)

	if err == nil {
		fmt.Println("告警规则ID已存在，更新规则")
		return update(client, rule, r)
	} else {
		folderName := fmt.Sprintf("%s-%s", ALERT_FOLDERTITLE, r.RuleID)
		folder, err := client.NewFolder(folderName)
		if err != nil {
			log.Fatalf("创建文件夹失败: %v", err)
		}
		datasource, _ := client.DataSourceByName("Prometheus")

		//3. bind notification(client)
		err = bindNotification(client)
		if err != nil {
			log.Fatalf("绑定通知失败: %v", err)
		}

		// 4. 创建告警规则
		ruleUID, err := client.NewAlertRule(newAlertRule(folder, *datasource, *r))
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
