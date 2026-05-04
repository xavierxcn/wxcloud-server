package db

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var dbInstance *gorm.DB
var dbMu sync.Mutex

// Init 初始化数据库
func Init() error {
	_, err := Ensure()
	return err
}

// Ensure returns a database connection, initializing it lazily on first use.
func Ensure() (*gorm.DB, error) {
	dbMu.Lock()
	defer dbMu.Unlock()

	if dbInstance != nil {
		return dbInstance, nil
	}

	user := os.Getenv("MYSQL_USERNAME")
	pwd := os.Getenv("MYSQL_PASSWORD")
	addr := os.Getenv("MYSQL_ADDRESS")
	dataBase := os.Getenv("MYSQL_DATABASE")
	if dataBase == "" {
		dataBase = "golang_demo"
	}
	if err := validateConfig(user, pwd, addr); err != nil {
		return nil, err
	}

	source := fmt.Sprintf("%s:%s@tcp(%s)/%s?readTimeout=1500ms&writeTimeout=1500ms&charset=utf8&loc=Local&parseTime=true", user, pwd, addr, dataBase)
	fmt.Printf("start init mysql with %s@tcp(%s)/%s\n", user, addr, dataBase)

	db, err := gorm.Open(mysql.Open(source), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // use singular table name, table for `User` would be `user` with this option enabled
		}})
	if err != nil {
		fmt.Println("DB Open error,err=", err.Error())
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("DB Init error,err=", err.Error())
		return nil, err
	}

	// 用于设置连接池中空闲连接的最大数量
	sqlDB.SetMaxIdleConns(100)
	// 设置打开数据库连接的最大数量
	sqlDB.SetMaxOpenConns(200)
	// 设置了连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(time.Hour)

	dbInstance = db

	fmt.Printf("finish init mysql with %s@tcp(%s)/%s\n", user, addr, dataBase)
	return dbInstance, nil
}

// Get ...
func Get() *gorm.DB {
	return dbInstance
}

func validateConfig(user, pwd, addr string) error {
	missing := make([]string, 0, 3)
	if strings.TrimSpace(addr) == "" {
		missing = append(missing, "MYSQL_ADDRESS")
	}
	if strings.TrimSpace(user) == "" {
		missing = append(missing, "MYSQL_USERNAME")
	}
	if strings.TrimSpace(pwd) == "" {
		missing = append(missing, "MYSQL_PASSWORD")
	}
	if len(missing) > 0 {
		return errors.New("missing database environment variables: " + strings.Join(missing, ", "))
	}

	return nil
}
