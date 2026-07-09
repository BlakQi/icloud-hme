// Command icloud-hme 启动 iCloud Hide My Email 多账号管理平台。
//
// 两个核心 HTTP 接口:
//
//	POST /api/create  — 创建隐私邮箱别名
//	GET  /api/inbox   — 读取邮件
//
// 用法:
//
//	./icloud-hme                    # 默认 :8081
//	./icloud-hme -addr :9000        # 指定端口
//	./icloud-hme -data ./data       # 指定数据目录
//	./icloud-hme -debug             # 调试模式
//	./icloud-hme -log-level debug   # 日志级别 (debug/info/warn/error)
package main

import (
	"flag"
	"path/filepath"

	"icloud-hme/internal/account"
	"icloud-hme/internal/log"
	"icloud-hme/internal/server"
)

func main() {
	addr := flag.String("addr", ":8081", "HTTP 监听地址")
	dataDir := flag.String("data", "./data", "数据目录 (accounts.json 存放位置)")
	logLevel := flag.String("log-level", "info", "日志级别: debug, info, warn, error")
	logFormat := flag.String("log-format", "console", "日志格式: console (彩色), json")
	debug := flag.Bool("debug", false, "调试模式 (启用详细日志)")
	flag.Parse()

	log.SetFormat(log.LogFormat(*logFormat))
	log.SetLevel(*logLevel)
	if *debug {
		log.SetLevel("debug")
	}

	log.Logger.Info().Str("addr", *addr).Msg("iCloud Hide My Email 服务启动")

	abs, err := filepath.Abs(*dataDir)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("数据目录路径错误")
	}

	mgr, err := account.NewManager(abs)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("初始化账号管理器失败")
	}
	count := len(mgr.ListAccounts())
	log.Logger.Info().Int("count", count).Str("data_dir", abs).Msg("账号加载完成")

	srv := server.New(mgr, *debug)

	log.Logger.Info().Str("addr", *addr).Msg("HTTP 服务就绪")
	if err := srv.Run(*addr); err != nil {
		log.Logger.Fatal().Err(err).Msg("服务启动失败")
	}
}
