package main

import (
	"context"

	"time"
	// 需要引入zap
	"go.uber.org/zap"
	// 引入该包
	// 或gitee.com/xiaoshanyangcode/yanglog
	logger "github.com/xiaoshanyangcode/yanglog"
)

func main() {
	// 全局退出函数
	ctx, cancel := context.WithCancel(context.Background())
	// 配置，只需要传LogConf这一个参数（子参数全部是可选的，可一个都不填）
	logconf := logger.LogConf{InfoFile: "looog/info_utc.log"}
	//  生成日志对象
	log := logger.NewLogger(ctx, logconf)
	defer log.Sync()

	// 可以正常使用了
	// 用法同zap的sugar，本身就是
	log.Error("This is an error message")
	log.Debugw("debug", "asdf", "asdfsdfaf")
	// 循环创建日志举例
	go func(log *zap.SugaredLogger) {
		for i := 0; i < 1000; i++ {
			log.Error("This is an error message")
			log.Debugw("debug", "asdf", "asdfsdfaf")
			time.Sleep(time.Millisecond * 1000)
		}
	}(log)
	time.Sleep(time.Minute)
	cancel()
}
