package logger

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type LokiHandler struct {
	url        string
	labels     map[string]string
	certPath   string
	httpClient *http.Client
}

type LokiPayload struct {
	Streams []LokiStream `json:"streams"`
}

type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

func NewLokiHandler(lokiURL, serviceName, certPath string) *LokiHandler {
	transport := &http.Transport{}

	if certPath == "" {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	client := &http.Client{
		Transport: transport,
	}

	return &LokiHandler{
		url:      lokiURL,
		certPath: certPath,
		labels: map[string]string{
			"job": serviceName,
			"app": "microservices",
		},
		httpClient: client,
	}
}

func (h *LokiHandler) Emit(level string, levelNumber int, message string) {
	logEntry := message
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())

	payload := LokiPayload{
		Streams: []LokiStream{
			{
				Stream: map[string]string{
					"job":          h.labels["job"],
					"app":          h.labels["app"],
					"level":        level,
					"level_number": fmt.Sprintf("%d", levelNumber),
				},
				Values: [][]string{
					{timestamp, logEntry},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("FAILED TO MARSHAL LOG DATA: %v\n", err)
		return
	}

	req, err := http.NewRequest(
		http.MethodPost,
		h.url,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Printf("FAILED TO CREATE REQUEST: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		fmt.Printf("FAILED TO SEND LOG TO LOKI. [EXCEPTION]: %v\n", err)
		fmt.Printf("log_data=%s\n", string(jsonData))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		fmt.Printf("Failed to send log to Loki: status=%d\n", resp.StatusCode)
	}
}

type Logger struct {
	logger      *log.Logger
	lokiHandler *LokiHandler
}

func SetupLogger(lokiURL, serviceName, certPath string) *Logger {
	handler := NewLokiHandler(
		lokiURL,
		serviceName,
		certPath,
	)

	return &Logger{
		logger:      log.New(os.Stdout, "", log.LstdFlags),
		lokiHandler: handler,
	}
}

func (l *Logger) Info(message string) {
	l.logger.Println("[INFO]", message)
	l.lokiHandler.Emit("INFO", 20, message)
}

func (l *Logger) Error(message string) {
	l.logger.Println("[ERROR]", message)
	l.lokiHandler.Emit("ERROR", 40, message)
}

func (l *Logger) Warning(message string) {
	l.logger.Println("[WARNING]", message)
	l.lokiHandler.Emit("WARNING", 30, message)
}

func (l *Logger) Debug(message string) {
	l.logger.Println("[DEBUG]", message)
	l.lokiHandler.Emit("DEBUG", 10, message)
}
