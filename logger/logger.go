package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/go-kratos/kratos/v2/log"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// Config 定义日志配置
// 该结构与业务 proto 解耦，可在不同项目中复用
type Config struct {
	Level         string
	Format        string
	Output        string
	FilePath      string
	MaxSize       int
	MaxAge        int
	MaxBackups    int
	Compress      bool
	EnableConsole bool
}

const (
	defaultFilePath   = "logs/app.log"
	defaultOutput     = "both"
	defaultFormat     = "text"
	defaultMaxSize    = 100
	defaultMaxAge     = 30
	defaultMaxBackups = 10
)

// multiWriter 实现 io.Writer 接口，同时写入多个目标
type multiWriter struct {
	writers []io.Writer
}

func (mw *multiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
	}
	return len(p), nil
}

// NewLogger 根据配置创建日志记录器
func NewLogger(cfg *Config) log.Logger {
	conf := applyDefaults(cfg)

	var logger log.Logger

	switch conf.Output {
	case "file":
		logger = createFileLogger(&conf)
	case "stderr":
		logger = log.NewStdLogger(os.Stderr)
	case "stdout":
		logger = log.NewStdLogger(os.Stdout)
	case "both":
		logger = createMultiLogger(conf.FilePath, true, &conf)
	default:
		logger = log.NewStdLogger(os.Stdout)
	}

	if conf.EnableConsole && conf.Output != "both" && conf.Output != "stdout" {
		logger = createMultiLogger(conf.FilePath, true, &conf)
	}

	return formatLogger(logger, conf.Format)
}

// NewHelper 创建日志助手
func NewHelper(cfg *Config) *log.Helper {
	return log.NewHelper(NewLogger(cfg))
}

func applyDefaults(cfg *Config) Config {
	if cfg == nil {
		return Config{
			Output:        defaultOutput,
			FilePath:      defaultFilePath,
			Format:        defaultFormat,
			MaxSize:       defaultMaxSize,
			MaxAge:        defaultMaxAge,
			MaxBackups:    defaultMaxBackups,
			Compress:      true,
			EnableConsole: true,
		}
	}

	c := *cfg
	if c.Output == "" {
		c.Output = defaultOutput
	}
	if c.FilePath == "" {
		c.FilePath = defaultFilePath
	}
	if c.Format == "" {
		c.Format = defaultFormat
	}
	if c.MaxSize == 0 {
		c.MaxSize = defaultMaxSize
	}
	if c.MaxAge == 0 {
		c.MaxAge = defaultMaxAge
	}
	if c.MaxBackups == 0 {
		c.MaxBackups = defaultMaxBackups
	}
	return c
}

func createFileLogger(cfg *Config) log.Logger {
	logDir := filepath.Dir(cfg.FilePath)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return log.NewStdLogger(os.Stdout)
	}

	fileWriter := &lumberjack.Logger{
		Filename:   cfg.FilePath,
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
		Compress:   cfg.Compress,
	}

	return log.NewStdLogger(fileWriter)
}

func createMultiLogger(filePath string, enableConsole bool, cfg *Config) log.Logger {
	logDir := filepath.Dir(filePath)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return log.NewStdLogger(os.Stdout)
	}

	fileWriter := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
		Compress:   cfg.Compress,
	}

	if enableConsole {
		multi := &multiWriter{writers: []io.Writer{fileWriter, os.Stdout}}
		return log.NewStdLogger(multi)
	}

	return log.NewStdLogger(fileWriter)
}

func formatLogger(logger log.Logger, format string) log.Logger {
	switch format {
	case "json":
		return log.With(logger,
			"ts", log.DefaultTimestamp,
			"caller", log.DefaultCaller,
		)
	default:
		return log.With(logger,
			"ts", log.DefaultTimestamp,
			"caller", log.DefaultCaller,
		)
	}
}
