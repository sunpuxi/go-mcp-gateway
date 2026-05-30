package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/sunpuxi/go-mcp-gateway/pkg/logger"
)

// NewMySQL 创建并测试 MySQL 数据库连接
func NewMySQL(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	logger.Info("数据库连接成功")
	return db, nil
}
