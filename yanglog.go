package yanglog

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 功能说明：
// 1.日志方案，指定info和error的日志文件
// 2.默认配置了：日志轮转,压缩，清理旧日志，可按需调整参数
// 3.用了zap的sugar功能
// 4.输出日志为json格式

// 问题点：
// 1.轮转日志的文件名的日期为utc时间。解决：文件名上加个utc字段

// 参考链接：
// 1.https://zhuanlan.zhihu.com/p/617294320?from_wecom=1
// 2.https://www.hanhandato.top/archives/golang-uber-go-zap?from_wecom=1

// 注意：轮转后旧的日志文件名自动带日期（UTC时间），建议在名字上带上utc。
type LogConf struct {
	// 注意：轮转后旧的日志文件名自动带日期（UTC时间），建议在名字上带上utc。
	InfoFile   string // 可选。请写绝对路径。默认值：可执行程序目录/log/info_utc.log
	ErrorFile  string // 可选。请写绝对路径。默认值：可执行程序目录/log/error_utc.log
	MaxSize    int    // 可选。MB。默认值：1000
	MaxBackups int    // 可选。保留旧日志文件的最大数量，默认值：1000
	MaxAge     int    // 可选。保留旧日志的最大天数，默认值：90
}

// 终端输出debug级别的日志，文件输出info和error日志。如果不满足要求，请自行修改
func NewLogger(ctx context.Context, config LogConf) *zap.SugaredLogger {
	// 终端输出debug级别的日志，文件输出info和error日志。如果不满足要求，请自行修改
	renderconf := config.render()
	// 创建一个Encoder配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, //CapitalLevelEncoder或LowercaseLevelEncoder 告警级别大写还是小写
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建一个写入器配置,Filename带目录(多级)的话，如果不存在会自动创建
	infofileConf := &lumberjack.Logger{
		Filename:   renderconf.InfoFile,
		MaxSize:    renderconf.MaxSize,    // MB
		MaxBackups: renderconf.MaxBackups, // 保留旧日志文件的最大数量。我们的程序默认值看renderLogConf函数
		MaxAge:     renderconf.MaxAge,     // 保留旧日志的最大天数
		Compress:   true,                  // disabled by default
	}
	infoFile := zapcore.AddSync(infofileConf)
	// 创建另一个写入器配置,Filename带目录(多级)的话，如果不存在会自动创建
	errorfileConf := &lumberjack.Logger{
		Filename:   renderconf.ErrorFile,
		MaxSize:    renderconf.MaxSize,    // MB
		MaxBackups: renderconf.MaxBackups, // 保留旧日志文件的最大数量。我们的程序默认值看renderLogConf函数
		MaxAge:     renderconf.MaxAge,     // 保留旧日志的最大天数
		Compress:   true,                  // disabled by default
	}
	errorFile := zapcore.AddSync(errorfileConf)

	// 创建一个核心配置
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), infoFile, zap.InfoLevel),
		// error级别及以上，单独写到指定文件
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), errorFile, zap.ErrorLevel),
		// 输出的终端
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
	)
	// 创建一个Logger实例
	logger := zap.New(core, zap.AddCaller()).Sugar()
	// newlog := Objlogger(*logger)

	// 程序启动时先轮转日志
	infofileConf.Rotate()
	errorfileConf.Rotate()

	// 每天0点生成新的日志文件文件
	go func(ctx context.Context) {
		for {
			now := time.Now()
			if now.Hour() == 0 && now.Minute() == 0 && now.Second() == 0 {
				infofileConf.Rotate()
				errorfileConf.Rotate()
			}
			// fmt.Println(now.Hour(), now.Minute(), now.Second())
			// 程序退出时，主动结束goroutine
			select {
			case <-ctx.Done(): // 等待上级通知
				goto endLoop
			default:
			}
			time.Sleep(1 * time.Second)
		}
	endLoop:
		// fmt.Println("exit cron goroutine")
	}(ctx)

	return logger
}

func getDir() string {
	// 获取当前可执行程序所在目录
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return exPath
}

func (config LogConf) render() LogConf {
	default_dir := getDir() + "/log/"
	// 没配置该参数时，设置默认值。默认的error日志文件
	if len(config.ErrorFile) == 0 {
		config.ErrorFile = default_dir + "error_utc.log"
	}
	// 没配置该参数时，设置默认值。默认的info日志文件
	if len(config.InfoFile) == 0 {
		config.InfoFile = default_dir + "info_utc.log"
	}
	// 没配置该参数时，设置默认值。默认的单日志文件最大大小，MB
	if config.MaxSize == 0 {
		config.MaxSize = 1000
	}
	// 没配置该参数时，设置默认值。默认的保留旧日志文件的最大数量
	if config.MaxBackups == 0 {
		config.MaxBackups = 1000
	}
	// 没配置该参数时，设置默认值。默认的保留旧日志的最大天数
	if config.MaxAge == 0 {
		config.MaxAge = 90
	}
	return config
}
