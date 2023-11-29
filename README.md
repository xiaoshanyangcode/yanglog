##  README
####  yanglog日志包，适合于golang语言
本模块，是在高性能日志包[zap](https://github.com/uber-go/zap/tree/master)和日志切割[lumberjack](https://github.com/natefinch/lumberjack)的基础上封装了一层。只需要传一个参数，即可完成配置。使用更加傻瓜，更加方便。

功能说明：

 1.可自定义指定info（包含error）和error的日志文件

 2.默认配置了：日志轮转，压缩，清理旧日志。可查看代码看默认参数值

 3.用了zap（高性能日志库）的sugar功能

 4.输出日志为json格式

####  使用代码举例:

备注：需要引入zap

```go
package main
import (
 "context"
 // 引入该包
 // 或github.com/xiaoshanyangcode/yanglog
 logger  "gitee.com/xiaoshanyangcode/yanglog"
 "time"
 // 需要引入zap
 "go.uber.org/zap" 
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
```