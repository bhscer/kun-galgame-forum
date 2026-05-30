package database

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"kun-galgame-api/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// withTimeZone pins the Postgres session TimeZone to Asia/Shanghai when the DSN
// doesn't already set one. Calendar-day logic (admin daily-stats date_trunc,
// the midnight reset cron) must agree on the day boundary; without a pinned
// zone, date_trunc('day', ...) buckets in the server's local zone (often UTC)
// while the Go side truncates in Asia/Shanghai → off-by-one oldest bucket.
func withTimeZone(dsn string) string {
	if strings.Contains(dsn, "TimeZone=") || strings.Contains(dsn, "timezone=") {
		return dsn
	}
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep + "TimeZone=Asia/Shanghai"
}

func NewPostgres(cfg config.DatabaseConfig, mode string) *gorm.DB {
	var logLevel logger.LogLevel
	if mode == "dev" {
		logLevel = logger.Info
	} else {
		logLevel = logger.Warn
	}

	db, err := gorm.Open(postgres.Open(withTimeZone(cfg.URL)), &gorm.Config{
		Logger:                 logger.Default.LogMode(logLevel),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic(fmt.Sprintf("连接数据库失败: %v", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("获取数据库连接池失败: %v", err))
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	slog.Info("数据库连接成功")
	return db
}
