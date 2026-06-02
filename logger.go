package logger

import (
	"log"
)

type Logger struct {
	URL      string
	Username string
	Password string
}

func New(url, username, password string) *Logger {
	return &Logger{
		URL:      url,
		Username: username,
		Password: password,
	}
}

func (l *Logger) Log(message string, level string) {
	// Здесь логика отправки в grafana
	log.Printf("[grafana %s] [Level %s]: %s \n",
		l.URL, level, message)
}
