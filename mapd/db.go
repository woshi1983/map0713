// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package mapd

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/logging"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	// sqldb is the global sqldb that is used by API Server.
	sqldb    *gorm.DB
	sqlMux   sync.RWMutex
	sqldbSet sync.Once
)

// InitDatabase attempts to setup global database connection for application.
func InitDatabase() (*gorm.DB, error) {
	var err error
	sqldbSet.Do(func() {
		sqlMux.Lock()
		defer sqlMux.Unlock()

		cfg := &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
			Logger: logger.Discard,
		}
		if GlobalConfig.DB.Debug {
			logCfg := logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: false,
			}
			if logging.IsColorfulTerm() {
				logCfg.Colorful = true
			}
			cfg.Logger = logger.New(log.New(os.Stderr, "\r\n", log.LstdFlags), logCfg)
		}

		// for mysql, the args must contains: `charset=utf8mb4&parseTime=true`
		sqldb, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", GlobalConfig.DB.User, GlobalConfig.DB.Password, GlobalConfig.DB.Host, GlobalConfig.DB.Name)), cfg)
		if err != nil {
			return
		}

		// set database connector runtime params.
		db, err := sqldb.DB()
		if err != nil {
			panic(fmt.Errorf("cannot access underlay database driver: %v", err))
		}
		db.SetMaxOpenConns(int(GlobalConfig.DB.MaxOpenConns))
		db.SetMaxIdleConns(int(GlobalConfig.DB.MaxIdleConns))
		db.SetConnMaxIdleTime(time.Minute)
	})
	if err != nil {
		return nil, fmt.Errorf("error establishing connection to database: %v", err)
	}
	return Database(), nil
}

// Database returns the current database connection.
func Database() *gorm.DB {
	sqlMux.RLock()
	defer sqlMux.RUnlock()

	return sqldb
}

// RegisterHandler is the handler for registering data model.
type RegisterHandler func(db *gorm.DB) error

// RegisterMiddleware is the middleware for registering data model.
type RegisterMiddleware func(next RegisterHandler) RegisterHandler

// RegisterModel registers data models and initializes the database.
func RegisterModel(db *gorm.DB, m ...RegisterMiddleware) error {
	f := func(db *gorm.DB) error { return nil }
	for i := len(m) - 1; i >= 0; i-- {
		f = m[i](f)
	}
	return f(db)
}
