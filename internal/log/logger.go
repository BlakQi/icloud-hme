// Package log 提供全局日志能力和 Gin 请求日志中间件。
//
// 基于 zerolog 实现,输出到 stderr,支持 JSON 和彩色终端两种格式。
package log

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// LogFormat 定义日志输出格式。
type LogFormat string

const (
	FormatConsole LogFormat = "console" // 彩色终端格式(默认)
	FormatJSON    LogFormat = "json"    // JSON 结构化格式
)

// Logger 是全局 zerolog 实例。
var Logger zerolog.Logger

func init() {
	SetFormat(FormatConsole)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// SetFormat 设置日志输出格式。
func SetFormat(f LogFormat) {
	if f == FormatJSON {
		Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	} else {
		Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
		}).With().Timestamp().Logger()
	}
}

// SetLevel 设置全局日志级别: debug, info, warn, error
func SetLevel(lvl string) {
	switch lvl {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// GinLogger 返回 Gin 请求日志中间件。
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		event := Logger.Info()
		if status >= 500 {
			event = Logger.Error()
		} else if status >= 400 {
			event = Logger.Warn()
		}

		if query != "" {
			event = event.Str("query", query)
		}

		event.
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Msg("request")
	}
}

// GinRecovery 返回 Gin 恢复中间件,捕获 panic 并记录详细错误。
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				Logger.Error().
					Interface("panic", r).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Msg("recovered from panic")
				c.AbortWithStatusJSON(500, gin.H{"success": false, "message": "服务器内部错误"})
			}
		}()
		c.Next()
	}
}
