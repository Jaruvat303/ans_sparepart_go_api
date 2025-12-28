package gormzap

import "gorm.io/gorm/logger"

func ParseLogLevel(s string) logger.LogLevel {
	switch s {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}
