package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"rufus-api/models/mqtt"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger(rmq *mqtt.RabbitMQ) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 記錄請求開始時間
		startTime := time.Now()

		// 讀取請求體
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = ioutil.ReadAll(c.Request.Body)
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))

		// 創建自定義ResponseWriter
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 處理請求
		c.Next()

		// 記錄響應時間
		endTime := time.Now()

		// 構造日誌消息
		logMessage := map[string]interface{}{
			"request": map[string]interface{}{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"body":   string(requestBody),
			},
			"response": map[string]interface{}{
				"status": c.Writer.Status(),
				"body":   blw.body.String(),
			},
			"duration": endTime.Sub(startTime).String(),
		}

		// 將日誌消息發送到RabbitMQ
		logJSON, _ := json.Marshal(logMessage)
		rmq.PublishMessage("logs", logJSON)
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}