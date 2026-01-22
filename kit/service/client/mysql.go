package client

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"sync"
	"time"
	_ "unitechs.com/cno/kit/db/dm"
	"unitechs.com/cno/kit/kctx"
)

// NewDBCli 创建一个mysl/dm 的db
/*
# 数据库类型  mysql/dm
Driver = "mysql"
Dsn = "root:root@tcp(localhost:3306)/ala?parseTime=true&loc=Local&"
MaxIdleConn = 10
MaxOpenConn = 10
#Dsn = "root:zyuc123@tcp(192.168.5.136:3306)/ala?parseTime=true&loc=Local&"
*/
func NewDBCli(ctx kctx.Ctx, getter IConfigGetter, prefix string) (*sql.DB, string, error) {

	dbType := getter.GetString(prefix + ".Driver")
	dsn := getter.GetString(prefix + ".Dsn")
	MaxOpenConn := getter.GetInt(prefix + ".MaxOpenConn")
	MaxIdleConn := getter.GetInt(prefix + ".MaxIdleConn")

	sDb, err := sql.Open(dbType, dsn)
	if err != nil {
		return nil, dbType, err
	}

	sDb.SetConnMaxLifetime(time.Second * 5)
	if MaxOpenConn == 0 {
		MaxOpenConn = 100
	}
	if MaxIdleConn == 0 {
		MaxIdleConn = 100
	}
	sDb.SetMaxIdleConns(MaxIdleConn)
	sDb.SetMaxOpenConns(MaxOpenConn)

	if err := sDb.Ping(); err != nil {
		return nil, dbType, errors.Wrapf(err, "ping")
	}

	return sDb, dbType, nil
}
func NewDBCliWithNacos(ctx kctx.Ctx, getter IConfigGetter, prefix string) (db *sql.DB, dbType string, err error) {

	if !getter.IsNacosSync() {
		ctx.Info("new db client without nacos hot update")
		return NewDBCli(ctx, getter, prefix)
	}
	configLst := []string{
		prefix + ".Driver",
		prefix + ".Dsn",
		prefix + ".MaxOpenConn",
		prefix + ".MaxIdleConn",
	}

	changed := make(chan struct{})

	for _, v := range configLst {
		func(v string) {
			if err_ := ListenNacos(getter, v, func(namespace, group, dataId, data string) {
				getter.Set(v, data)
				changed <- struct{}{}
			}); err_ != nil {
				ctx.Error("ListenNacos failed %v", err_)
			}
		}(v)
	}
	wg := sync.WaitGroup{}
	once := sync.Once{}
	wg.Add(1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-changed:
				if !ok {
					return
				}
				var db2 *sql.DB
				if db != nil {
					db2 = db
					db.Close()
				}

				db_, dbType_, err := NewDBCli(ctx, getter, prefix)
				if err != nil {
					ctx.Errorf("NewDBCli failed: %v", err)
					continue
				}
				db = db_
				dbType = dbType_

				if db2 != nil {
					db2.Close()
				}
				once.Do(func() {
					wg.Done()
				})
			}
		}
	}()

	changed <- struct{}{}
	wg.Wait()
	return
}
