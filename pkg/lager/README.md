#paas_lager
===========
A golang logging library for PaaS

paas_lager 是 PaaS运维组对 CloudFoundry 的 lager 库进行了封装和完善，目的是提供统一的接口方便 PaaS 自研模块的日志输出需求。
运维组对原始的 lager 日志输出中不符合公司日志要求规范的部分也做了加强和信息补全，以更方便实际问题的诊断和定位。补充的信息包括：进程ID，文件名，行号，方法名等。
使用 paas_lager 记录日志，能保证自研模块的日志输出为统一的 json 格式，方便后续日志在 Elasticsearch 中做进一步检索和分析。

##Usage
Download 
```
git clone http://rnd-isourceb.huawei.com/PaaS/paas-sdk.git $GOPATH/src/paas-sdk
```

Create logger (complex mode)
```
paas_lager.Init(paas_lager.Config{
        LoggerLevel:   loggerLevel,
        LoggerFile:    loggerFile,
        EnableRsyslog: enableRsyslog,
        RsyslogNetwork: "udp",
        RsyslogAddr:    "127.0.0.1:514",
        LogFormatText:  false,
})

logger := paas_lager.NewLogger(component)
```

* LoggerLevel: 日志级别由低到高分别为 DEBUG, INFO, WARN, ERROR, FATAL 共5个级别，这里设置的级别是日志输出的最低级别，只有不低于该级别的日志才会输出。所以，如果不想打印 WARN 以下级别的日志，可以把 LoggerLevel 设置为 WARN。
从减小日志大小的考虑，建议在单元测试的时候，可以把 LoggerLevel 级别设置的比较低，比如 DEBUG，而在正式部署的系统代码中，建议级别设置为 WARN，以减少日志频繁输出带来的高 IO 对 VM 和 PaaS 日志后台的存储和处理的负担。
* LoggerFile: 输出日志的文件名，为空则输出到 os.Stdout。建议路径统一放置在 /var/paas/sys/log 目录下，各个模块分别创建各自的子目录，这样后续可以通过 logrotate 统一进行日志的 rotate 防爆处理。
* EnableRsyslog: 设定是否输出日志到 rsyslog，默认为 false。
* RsyslogNetwork: 设定日志传输的方式，是 "udp" 还是 "tcp"，默认为"udp"
* RsyslogAddr: 设定日志发送的目的syslog地址
* LogFormatText: 设定日志的输出格式是 json 还是 plaintext (类似于log4j)，默认为 false，不建议修改，如果开发过程中想本地查看日志的话，可以设定 LoggerFile 和 LogFormatText 为 true，这样会输出类似于 log4j 格式的本地日志。但是建议正式的生产环境还是用 json 的格式输出。
* component: 日志中输出的模块名字，source 字段所指定的值

对于大多数的使用场景，可以只配置下面的参数即可。
Create logger (simple mode)
```
paas_lager.Init(paas_lager.Config{
        LoggerLevel:   loggerLevel,
        LoggerFile:    loggerFile,
        EnableRsyslog: enableRsyslog,
})

logger := paas_lager.NewLogger(component)
```

##Example
```
package main

import (
	"fmt"

	"paas_lager"
	"paas_lager/lager"
)

func main() {
	paas_lager.Init(paas_lager.Config{
		LoggerLevel:   "WARN",
		LoggerFile:    "/var/paas/sys/log/ops-mgr/example.log",
		EnableRsyslog: false,
	})

	logger := paas_lager.NewLogger("example")

	logger.InfoF("Hi %s, system is starting up ...", "paas-bot")

	logger.Debug("check-info", lager.Data{
		"info": "something",
	})

	err := fmt.Errorf("Oops, error occurred!")
	logger.Warn("failed-to-do-somthing", err, lager.Data{
		"info": "something",
	})

	err = fmt.Errorf("This is an error")
	logger.Error("failed-to-do-somthing", err)

	logger.Info("shutting-down")
}

```

The output is:
```
{"timestamp":"2015-11-09T14:19:16.825759987Z","source":"example","message":"example.Hi paas-bot, system is starting up ...","log_level":"info","data":{},"process_id":5234,"file":"paas_lager/examples/main.go","lineno":20,"method":"main"}{"timestamp":"2015-11-09T14:19:16.826463479Z","source":"example","message":"example.check-info","log_level":"debug","data":{"info":"something"},"process_id":5234,"file":"paas_lager/examples/main.go","lineno":24,"method":"main"}{"timestamp":"2015-11-09T14:19:16.826801265Z","source":"example","message":"example.failed-to-do-somthing","log_level":"warn","data":{"error":"Oops, error occurred!","info":"something"},"process_id":5234,"file":"paas_lager/examples/main.go","lineno":29,"method":"main"}
{"timestamp":"2015-11-09T14:19:16.827071328Z","source":"example","message":"example.failed-to-do-somthing","log_level":"error","data":{"error":"This is an error"},"process_id":5234,"file":"paas_lager/examples/main.go","lineno":32,"method":"main"}
{"timestamp":"2015-11-09T14:19:16.82732567Z","source":"example","message":"example.shutting-down","log_level":"info","data":{},"process_id":5234,"file":"paas_lager/examples/main.go","lineno":34,"method":"main"}
```


如果上述 paas_lager.Init 初始化部分，把 "LogFormatText: true," 加上的话，日志的输出内容会变成如下的 plaintext 模式，方便开发人员本地调试查看日志：
```
2015-11-09T14:19:53.73768695Z INFO example 5748 paas_lager/examples/main.go main():20 - example.Hi paas-bot, system is starting up ...
2015-11-09T14:19:53.73784007Z DEBUG example 5748 paas_lager/examples/main.go main():24 - example.check-info
2015-11-09T14:19:53.737859147Z WARN example 5748 paas_lager/examples/main.go main():29 - example.failed-to-do-somthing
2015-11-09T14:19:53.738067943Z ERROR example 5748 paas_lager/examples/main.go main():32 - example.failed-to-do-somthing
2015-11-09T14:19:53.738092109Z INFO example 5748 paas_lager/examples/main.go main():34 - example.shutting-down
```

## Suggestion to log4go, glog 
使用其他日志模块的，如果不方便迁移到 paas_lager，也建议参考上面的日志输出，把日志格式一致起来，转换为有语义的格式化文档。这样后续在 ElasticSearch 中，可以很方便的解析 json 日志，进行检索。
