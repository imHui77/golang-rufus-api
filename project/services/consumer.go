package services

import (
	"encoding/json"
	"log"
	"rufus-api/models/mqtt"
)

type LogMessage struct {
	Request  RequestInfo  `json:"request"`
	Response ResponseInfo `json:"response"`
	Duration string       `json:"duration"`
}

type RequestInfo struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Body   string `json:"body"`
}

type ResponseInfo struct {
	Status int    `json:"status"`
	Body   string `json:"body"`
}

func StartLogConsumer(rmq *mqtt.RabbitMQ) error {
	return rmq.ConsumeMessages("logs", processLogMessage)
}

func processLogMessage(body []byte) error {
	var logMessage LogMessage
	err := json.Unmarshal(body, &logMessage)
	if err != nil {
		return err
	}

	// 這裡可以根據需求處理日誌消息
	// 例如，將其保存到數據庫或發送到監控系統
	log.Printf("Received log: Method=%s, Path=%s, Status=%d, Duration=%s",
		logMessage.Request.Method,
		logMessage.Request.Path,
		logMessage.Response.Status,
		logMessage.Duration)

	return nil
}