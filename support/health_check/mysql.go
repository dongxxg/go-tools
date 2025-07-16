package hc

import (
	"context"
	"time"
)

const mysqlCheckDeadline = 15 * time.Second

type iPing interface {
	PingContext(ctx context.Context) error
}

func CheckMysql(db iPing) error {
	ctx, _ := context.WithTimeout(context.Background(), mysqlCheckDeadline)
	return db.PingContext(ctx)
}
