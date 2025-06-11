package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	grafana "github.com/grafana/grafana-api-golang-client"
)

var Client *grafana.Client

// 定义请求参数结构体
type AlertRequest struct {
	RuleID          string              `form:"ruleId" json:"rule_id"`
	MetricCode      string              `form:"metricCode" json:"metric_code"`
	Operator        string              `form:"operator" json:"operator"`
	MetricThreshold string              `form:"metricThreshold" json:"metric_threshold"`
	ContactList     []map[string]string `form:"contactList" json:"contact_list"`
}

func updateAlertRule2Http(c *gin.Context) {
	// 解析post 请求携带的参数

	var req AlertRequest
	var body map[string]any

	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("解析请求参数失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析请求参数失败: " + err.Error()})
		return
	}

	req.RuleID = body["ruleId"].(string)
	req.MetricCode = body["metricCode"].(string)
	req.Operator = body["operator"].(string)
	req.MetricThreshold = body["metricThreshold"].(string)
	contactList, ok := body["contactList"].([]interface{})
	if !ok {
		log.Printf("无效的contactList参数: %v", body["contactList"])
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的contactList参数"})
		return
	}
	// 将contactList转换为map[string]string切片
	contactMapList := make([]map[string]string, len(contactList))
	for i, item := range contactList {
		if itemMap, ok := item.(map[string]interface{}); ok {
			contactMap := make(map[string]string)
			for k, v := range itemMap {
				if strValue, ok := v.(string); ok {
					contactMap[k] = strValue
				} else {
					log.Printf("无效的contactList项: %v", v)
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的contactList项: " + k})
					return
				}
			}
			contactMapList[i] = contactMap
		} else {
			log.Printf("无效的contactList项: %v", item)
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的contactList项"})
			return
		}
	}
	req.ContactList = contactMapList

	// // 调用业务逻辑处理函数（这里只是示例）
	_, err := updateAlertRule(Client, &req)
	if err != nil {
		log.Printf("更新告警规则失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新告警规则失败: " + err.Error()})
		return
	} else {
		// 返回成功响应
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "告警规则已更新",
		})
	}
}

func main() {
	// WEBHOOK_URL 从命令行参数读取
	if len(os.Args) < 2 {
		log.Fatal("请提供 WEBHOOK_URL 参数")
	}
	WEBHOOK_URL = os.Args[1]

	Client, _ = createGrafanaClient()

	// 确保Webhook通知渠道存在
	if _, err := ensureWebhookContactPoint(Client); err != nil {
		log.Printf("创建通知渠道失败: %v", err)
	} else {
		log.Println("Webhook通知渠道已配置")
	}

	r := gin.Default()

	// 注册路由和处理函数
	r.POST("/api/v1/alert", updateAlertRule2Http)

	// 启动服务
	r.Run(":8080")
}

// func main() {
// 	client, err := createGrafanaClient()
// 	if err != nil {
// 		log.Fatalf("创建客户端失败: %v", err)
// 	}

// 	// 确保Webhook通知渠道存在
// 	if _, err := ensureWebhookContactPoint(client); err != nil {
// 		log.Printf("创建通知渠道失败: %v", err)
// 	} else {
// 		log.Println("Webhook通知渠道已配置")
// 	}

// 	if len(os.Args) < 2 {
// 		printHelp()
// 		return
// 	}

// 	switch os.Args[1] {
// 	case "list-rules":
// 		rules, err := listAlertRules()
// 		if err != nil {
// 			log.Fatalf("获取告警规则失败: %v", err)
// 		} else {
// 			for _, rule := range rules {
// 				dataStr := ""
// 				if rule.Data != nil {
// 					if b, err := json.Marshal(rule.Data); err == nil {
// 						dataStr = string(b)
// 					} else {
// 						dataStr = fmt.Sprintf("marshal error: %v", err)
// 					}
// 				}
// 				fmt.Printf("UID: %s, Name: %s, Group: %s, Folder: %s, Data: %s\n",
// 					rule.UID, rule.Title, rule.RuleGroup, rule.FolderUID, dataStr)
// 			}
// 		}
// 	case "update-rule":
// 		if len(os.Args) < 5 {
// 			log.Fatal("缺少参数: update-rule <folder> <group> <title> <expr> [duration]")
// 		}

// 		uid, err := updateAlertRule(
// 			client,
// 			os.Args[2],
// 			os.Args[3],
// 			os.Args[4],
// 		)
// 		if err != nil {
// 			log.Fatalf("创建规则失败: %v", err)
// 		}
// 		log.Printf("告警规则创建成功: UID=%s", uid)

// 	default:
// 		log.Fatalf("未知命令: %s", os.Args[1])
// 	}
// }

// func printHelp() {
// 	log.Println(`Grafana 告警管理工具 (使用 grafana-openapi-client-go)

// 命令:
//   update-rule cpu_threshold mem_threshold disk_threshold - 创建告警规则

// 示例:
//   # 创建CPU使用率告警
//   ./alerts list-rules
//   ./alerts update-rule 90 90 90`)
// }
