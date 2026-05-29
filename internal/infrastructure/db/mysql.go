package db

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// NewMySQL 创建并测试 MySQL 数据库连接
func NewMySQL(dsn string) *sqlx.DB {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Printf("数据库连接成功")
	return db
}
