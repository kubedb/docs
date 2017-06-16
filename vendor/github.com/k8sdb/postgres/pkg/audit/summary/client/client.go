package client

import (
	"fmt"
	"time"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

func NewEngine(username, password, host, port, dbName string) (*xorm.Engine, error) {
	cnnstr := fmt.Sprintf("user=%v password=%v host=%v port=%v dbname=%v sslmode=disable",
		username, password, host, port, dbName)

	engine, err := xorm.NewEngine("postgres", cnnstr)
	if err != nil {
		return nil, err
	}

	engine.SetMaxIdleConns(0)
	engine.DB().SetConnMaxLifetime(10 * time.Minute)
	engine.ShowSQL(false)
	engine.Logger().SetLevel(core.LOG_ERR)
	return engine, nil
}
