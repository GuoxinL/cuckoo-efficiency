package logger

import (
	logv2 "chainmaker.org/chainmaker/logger/v2"
	api "chainmaker.org/chainmaker/protocol/v2"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LOG_LEVEL 日志级别，int类型，内部接口使用常量
type LOG_LEVEL int

// 日志级别
const (
	LEVEL_DEBUG LOG_LEVEL = iota
	LEVEL_INFO
	LEVEL_WARN
	LEVEL_ERROR
)

type LogCfg struct {
	Module       string `mapstructure:"module"`         // module: module name
	ChainId      string `mapstructure:"chain_id"`       // chainId: chain id
	LogPath      string `mapstructure:"log_path"`       // logPath: log file save path
	LogLevel     string `mapstructure:"log_level"`      // logLevel: log level
	MaxAge       int    `mapstructure:"max_age"`        // maxAge: the maximum number of days to retain old log files
	RotationTime int    `mapstructure:"rotation_time"`  // RotationTime: rotation time
	RotationSize int64  `mapstructure:"rotation_size"`  // RotationSize: rotation size Mb
	JsonFormat   bool   `mapstructure:"json_format"`    // jsonFormat: log file use json format
	ShowLine     bool   `mapstructure:"show_line"`      // showLine: show filename and line number
	LogInConsole bool   `mapstructure:"log_in_console"` // logInConsole: show logs in console at the same time
	ShowColor    bool   `mapstructure:"show_color"`     // if true, show color log
	IsBrief      bool   `mapstructure:"is_brief"`       // if true, only show log, won't print log level、caller func and line

	// StackTraceLevel record a stack trace for all messages at or above a given level.
	// Empty string or invalid level will not open stack trace.
	StackTraceLevel string
}

func GetLogger0(config LogCfg) *zap.SugaredLogger {
	conf := &LogConfig{
		Module:          config.Module,
		ChainId:         config.ChainId,
		LogPath:         config.LogPath,
		LogLevel:        GetLogLevel(config.LogLevel),
		MaxAge:          config.MaxAge,
		RotationTime:    config.RotationTime,
		RotationSize:    config.RotationSize,
		JsonFormat:      config.JsonFormat,
		ShowLine:        config.ShowLine,
		LogInConsole:    config.LogInConsole,
		ShowColor:       config.ShowColor,
		IsBrief:         config.IsBrief,
		StackTraceLevel: config.StackTraceLevel,
	}
	logger, _ := InitSugarLogger(conf)
	return logger
}

func GetLogger1(config LogCfg) (api.Logger, *zap.SugaredLogger) {

	mLog := make(map[string]logv2.LogNodeConfig)
	conf := &logv2.LogConfig{
		SystemLog: logv2.LogNodeConfig{
			LogLevelDefault: config.LogLevel,
			FilePath:        config.LogPath,
			MaxAge:          config.MaxAge,
			RotationTime:    config.RotationTime,
			RotationSize:    config.RotationSize,
			LogInConsole:    config.LogInConsole,
			LogByChain:      true,
			JsonFormat:      config.JsonFormat,
			ShowColor:       config.ShowColor,
			StackTraceLevel: config.StackTraceLevel,
		},
		ModuleLog: mLog,
	}
	logv2.SetLogConfig(conf)

	logger := logv2.GetLoggerByChain(config.Module, config.ChainId)
	return logger, logger.Logger()
}

// 日志级别，配置文件定义的常量
const (
	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
)

// GetLogLevel 根据字符串型的日志级别，返回枚举型日志级别
// @param lvl
// @return LOG_LEVEL
func GetLogLevel(lvl string) LOG_LEVEL {
	switch strings.ToUpper(lvl) {
	case ERROR:
		return LEVEL_ERROR
	case WARN:
		return LEVEL_WARN
	case INFO:
		return LEVEL_INFO
	case DEBUG:
		return LEVEL_DEBUG
	}

	return LEVEL_INFO
}

func getZapLevel(lvl string) (*zapcore.Level, error) {
	var zapLevel zapcore.Level
	switch strings.ToUpper(lvl) {
	case ERROR:
		zapLevel = zap.ErrorLevel
	case WARN:
		zapLevel = zap.WarnLevel
	case INFO:
		zapLevel = zap.InfoLevel
	case DEBUG:
		zapLevel = zap.DebugLevel
	default:
		return nil, errors.New("invalid log level")
	}
	return &zapLevel, nil
}

// 日志切割默认配置
const (
	DEFAULT_MAX_AGE       = 365 // 日志最长保存时间，单位：天
	DEFAULT_ROTATION_TIME = 6   // 日志滚动间隔，单位：小时
	DEFAULT_ROTATION_SIZE = 100 // 默认的日志滚动大小，单位：MB
)

//日志滚动单位
const (
	ROTATION_SIZE_MB = 1024 * 1024
)

type color int

// 常用颜色
const (
	ColorBlack color = iota + 30
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

var colorList = [...]color{ColorRed, ColorGreen, ColorYellow, ColorBlue, ColorMagenta}

var hookMap = make(map[string]struct{})

// LogConfig 日志记录的配置
type LogConfig struct {
	Module       string    // module: module name
	ChainId      string    // chainId: chain id
	LogPath      string    // logPath: log file save path
	LogLevel     LOG_LEVEL // logLevel: log level
	MaxAge       int       // maxAge: the maximum number of days to retain old log files
	RotationTime int       // RotationTime: rotation time
	RotationSize int64     // RotationSize: rotation size Mb
	JsonFormat   bool      // jsonFormat: log file use json format
	ShowLine     bool      // showLine: show filename and line number
	LogInConsole bool      // logInConsole: show logs in console at the same time
	ShowColor    bool      // if true, show color log
	IsBrief      bool      // if true, only show log, won't print log level、caller func and line

	// StackTraceLevel record a stack trace for all messages at or above a given level.
	// Empty string or invalid level will not open stack trace.
	StackTraceLevel string
}

// InitSugarLogger 基于配置初始化一个zap的SugaredLogger
// @param logConfig
// @param writer
// @return *zap.SugaredLogger
// @return zap.AtomicLevel
func InitSugarLogger(logConfig *LogConfig, writer ...io.Writer) (*zap.SugaredLogger, zap.AtomicLevel) {
	var level zapcore.Level
	switch logConfig.LogLevel {
	case LEVEL_DEBUG:
		level = zap.DebugLevel
	case LEVEL_INFO:
		level = zap.InfoLevel
	case LEVEL_WARN:
		level = zap.WarnLevel
	case LEVEL_ERROR:
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	aLevel := zap.NewAtomicLevel()
	aLevel.SetLevel(level)

	sugaredLogger := newLogger(logConfig, aLevel, writer...).Sugar()

	return sugaredLogger, aLevel
}

func newLogger(logConfig *LogConfig, level zap.AtomicLevel, writer ...io.Writer) *zap.Logger {
	var (
		hook io.Writer
		ok   bool
		err  error
	)

	_, ok = hookMap[logConfig.LogPath]
	if !ok {
		hook, err = getHook(logConfig.LogPath, logConfig.MaxAge, logConfig.RotationTime, logConfig.RotationSize)
		if err != nil {
			log.Fatalf("new logger get hook failed, %s", err)
		}
		hookMap[logConfig.LogPath] = struct{}{}
	} else {
		hook, err = getHook(logConfig.LogPath, logConfig.MaxAge, 0, logConfig.RotationSize)
		if err != nil {
			log.Fatalf("new logger get hook failed, %s", err)
		}
	}

	var syncer zapcore.WriteSyncer
	syncers := []zapcore.WriteSyncer{zapcore.AddSync(hook)}
	if logConfig.LogInConsole {
		syncers = append(syncers, zapcore.AddSync(os.Stdout))
	}
	for _, outSyncer := range writer {
		syncers = append(syncers, zapcore.AddSync(outSyncer))
	}

	syncer = zapcore.NewMultiWriteSyncer(syncers...)

	var encoderConfig zapcore.EncoderConfig
	if logConfig.IsBrief {
		encoderConfig = zapcore.EncoderConfig{
			TimeKey:    "time",
			MessageKey: "msg",
			EncodeTime: CustomTimeEncoder,
			LineEnding: zapcore.DefaultLineEnding,
		}
	} else {
		encoderConfig = zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "line",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    CustomLevelEncoder,
			EncodeTime:     CustomTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		}
	}

	var encoder zapcore.Encoder
	if logConfig.JsonFormat {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		syncer,
		level,
	)

	chainId := fmt.Sprintf("@%s", logConfig.ChainId)
	if logConfig.ShowColor {
		chainId = getColorChainId(chainId)
	}

	var name string
	if logConfig.ChainId != "" {
		name = fmt.Sprintf("%s %s", logConfig.Module, chainId)
	} else {
		name = logConfig.Module
	}

	logger := zap.New(core).Named(name)
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	if logConfig.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}
	if lvl, err := getZapLevel(logConfig.StackTraceLevel); err == nil {
		logger = logger.WithOptions(zap.AddStacktrace(lvl))
	}
	logger = logger.WithOptions(zap.AddCallerSkip(0))
	return logger
}

func getHook(filename string, maxAge, rotationTime int, rotationSize int64) (io.Writer, error) {

	hook, err := rotatelogs.New(
		filename+".%Y%m%d%H",
		rotatelogs.WithRotationTime(time.Hour*time.Duration(rotationTime)),
		//filename+".%Y%m%d%H%M",
		//rotatelogs.WithRotationSize(rotationSize*ROTATION_SIZE_MB),
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*24*time.Duration(maxAge)),
	)

	if err != nil {
		return nil, err
	}

	return hook, nil
}

// nolint: deadcode, unused
func recognizeLogLevel(l string) LOG_LEVEL {
	logLevel := strings.ToUpper(l)
	var level LOG_LEVEL
	switch logLevel {
	case DEBUG:
		level = LEVEL_DEBUG
	case INFO:
		level = LEVEL_INFO
	case WARN:
		level = LEVEL_WARN
	case ERROR:
		level = LEVEL_ERROR
	default:
		level = LEVEL_INFO
	}
	return level
}

// CustomLevelEncoder 自定义日志级别的输出格式
// @param level
// @param enc
func CustomLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}

// CustomTimeEncoder 自定义时间转字符串的编码方法
// @param t
// @param enc
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// nolint: deadcode, unused
func showColor(color color, msg string) string {
	return fmt.Sprintf("\033[%dm%s\033[0m", int(color), msg)
}

func showColorBold(color color, msg string) string {
	return fmt.Sprintf("\033[%d;1m%s\033[0m", int(color), msg)
}

func getColorChainId(chainId string) string {
	c := crc32.ChecksumIEEE([]byte(chainId))
	color := colorList[int(c)%len(colorList)]
	return showColorBold(color, chainId)
}
