package log

import (
	"log"
	"os"
)

type Log struct {
	Error *log.Logger
	Info  *log.Logger
}

func NewLogger() *Log {
	return &Log{
		Error: log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		Info:  log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
	}
}
