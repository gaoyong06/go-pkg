package logger

// Package logger 提供基于 Kratos 框架的统一日志处理功能。
// 注意：本包专为 Kratos 微服务设计，依赖 Kratos 的 log 接口。
// 对于非 Kratos 项目（如 Gin 框架的 schedule_manager），请勿使用本包，
// 应直接在项目内部实现日志逻辑，以避免不必要的依赖和复杂度。

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// Config 日志配置结构体
// 调用者需将具体的配置对象映射到此结构体
type Config struct {
	Level         string // 日志级别: debug, info, warn, error
	Format        string // 日志格式: json, text
	Output        string // 输出位置: stdout, stderr, file, both
	FilePath      string // 日志文件路径 (当 Output 为 file 或 both 时有效)
	MaxSize       int    // 单个日志文件最大大小 (MB)
	MaxAge        int    // 日志文件保留天数
	MaxBackups    int    // 日志文件最大备份数量
	Compress      bool   // 是否压缩旧日志
	EnableConsole bool   // 是否同时输出到控制台 (辅助字段，通常由 Output 决定)
}

const (
	defaultFilePath   = "logs/app.log"
	defaultOutput     = "both"
	defaultFormat     = "text"
	defaultMaxSize    = 100
	defaultMaxAge     = 30
	defaultMaxBackups = 10
)

// InitLogger 初始化日志记录器 (Kratos 专用)
// 设计原则：简单至上，仅满足 Kratos 微服务的标准日志需求
//
// 参数:
//
//	c: 日志配置对象 (如果为 nil，将使用默认配置)
//	id, name, version: 服务标识信息
//
// 返回:
//
//	log.Logger: 初始化后的日志记录器
//	string: 日志文件路径 (供启动日志使用)
func InitLogger(c *Config, id, name, version string) (log.Logger, string) {
	// 1. 应用默认值
	conf := applyDefaults(c)

	// 2. 创建 Logger (Zap 实现)
	loggerInstance := newLogger(conf)

	// 3. 添加基本字段
	// 注意: 我们保留 Kratos 的 DefaultTimestamp 和 DefaultCaller 中间件
	// 因此在 Zap 配置中禁用了内置的时间和 Caller，避免重复
	return log.With(loggerInstance,
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", name,
		"service.version", version,
	), conf.FilePath
}

// ZapLogger 适配 Kratos 的 Logger 接口
type ZapLogger struct {
	log *zap.Logger
}

// Log 实现 log.Logger 接口
func (l *ZapLogger) Log(level log.Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 {
		return nil
	}
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "")
	}

	var msg string
	fields := make([]zap.Field, 0, len(keyvals)/2)

	for i := 0; i < len(keyvals); i += 2 {
		key := fmt.Sprintf("%v", keyvals[i])
		// 提取 msg 字段作为 Zap 的主消息
		if key == "msg" {
			msg = fmt.Sprintf("%v", keyvals[i+1])
			continue
		}
		fields = append(fields, zap.Any(key, keyvals[i+1]))
	}

	switch level {
	case log.LevelDebug:
		l.log.Debug(msg, fields...)
	case log.LevelInfo:
		l.log.Info(msg, fields...)
	case log.LevelWarn:
		l.log.Warn(msg, fields...)
	case log.LevelError:
		l.log.Error(msg, fields...)
	case log.LevelFatal:
		l.log.Fatal(msg, fields...)
	}
	return nil
}

// Sync 同步缓冲区
func (l *ZapLogger) Sync() error {
	return l.log.Sync()
}

// ======================================== private method ========================================
func applyDefaults(c *Config) *Config {
	if c == nil {
		return &Config{
			Output:     defaultOutput,
			FilePath:   defaultFilePath,
			Format:     defaultFormat,
			MaxSize:    defaultMaxSize,
			MaxAge:     defaultMaxAge,
			MaxBackups: defaultMaxBackups,
			Compress:   true,
		}
	}

	// 复制一份配置，避免修改原对象
	conf := *c

	if conf.Output == "" {
		conf.Output = defaultOutput
	}
	if conf.FilePath == "" {
		conf.FilePath = defaultFilePath
	}
	if conf.Format == "" {
		conf.Format = defaultFormat
	}
	if conf.MaxSize == 0 {
		conf.MaxSize = defaultMaxSize
	}
	if conf.MaxAge == 0 {
		conf.MaxAge = defaultMaxAge
	}
	if conf.MaxBackups == 0 {
		conf.MaxBackups = defaultMaxBackups
	}
	return &conf
}

// newLogger 根据配置创建日志记录器 (私有方法)
func newLogger(conf *Config) log.Logger {
	// 1. 配置 WriteSyncer (输出位置)
	var writeSyncer zapcore.WriteSyncer

	// 确保日志目录存在
	if conf.Output == "file" || conf.Output == "both" {
		logDir := filepath.Dir(conf.FilePath)
		_ = os.MkdirAll(logDir, 0o755)
	}

	fileWriter := &lumberjack.Logger{
		Filename:   conf.FilePath,
		MaxSize:    conf.MaxSize,
		MaxAge:     conf.MaxAge,
		MaxBackups: conf.MaxBackups,
		Compress:   conf.Compress,
	}

	switch conf.Output {
	case "file":
		writeSyncer = zapcore.AddSync(fileWriter)
	case "stderr":
		writeSyncer = zapcore.AddSync(os.Stderr)
	case "stdout":
		writeSyncer = zapcore.AddSync(os.Stdout)
	case "both":
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileWriter), zapcore.AddSync(os.Stdout))
	default:
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// 2. 配置 Encoder (日志格式)
	encoderConfig := zap.NewProductionEncoderConfig()
	// 使用 ISO8601 时间格式 (虽然我们下面禁用了 TimeKey，但保留此配置以防未来启用)
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 禁用 Zap 自带的 TimeKey 和 CallerKey，因为 Kratos 中间件已经提供了 "ts" 和 "caller"
	// 这样可以避免日志中出现重复字段
	encoderConfig.TimeKey = ""
	encoderConfig.CallerKey = ""

	var encoder zapcore.Encoder
	if conf.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		// Console 格式适合开发环境
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 3. 配置 Level (日志级别)
	var zapLevel zapcore.Level
	if conf.Level == "" {
		zapLevel = zapcore.InfoLevel
	} else {
		if err := zapLevel.UnmarshalText([]byte(conf.Level)); err != nil {
			zapLevel = zapcore.InfoLevel
		}
	}

	// 4. 创建 Core 和 Logger
	core := zapcore.NewCore(encoder, writeSyncer, zapLevel)
	zapLog := zap.New(core)

	return &ZapLogger{log: zapLog}
}
