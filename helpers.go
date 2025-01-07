package health

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

var (
	_ Checker        = (*namedFunc)(nil)
	_ slog.LogValuer = (*namedFunc)(nil)
)

func RegisterSQLDB(
	ctx context.Context,
	db *sql.DB,
	name string,
	base Check,
) Registered {
	return registerPinger(ctx, db, "sql.DB", name, base)
}

func RegisterSQLConn(
	ctx context.Context,
	conn *sql.Conn,
	name string,
	base Check,
) Registered {
	return registerPinger(ctx, conn, "sql.Conn", name, base)
}

func registerPinger(
	ctx context.Context,
	p interface {
		PingContext(context.Context) error
	},
	n,
	name string,
	base Check,
) Registered {
	if base.ComponentType == "" {
		base.ComponentType = ComponentTypeDatastore
	}

	return Register(ctx, name, namedFunc{
		string: n,
		CheckerFunc: func(ctx context.Context) []Check {
			c := base

			start := time.Now()
			err := p.PingContext(ctx)
			c.SetObservedTime(time.Since(start))

			if err != nil {
				c.Status = StatusFail
				c.Output = err.Error()
			}

			return []Check{c}
		},
	})
}

type namedFunc struct {
	string
	CheckerFunc
}

func (f namedFunc) LogValue() slog.Value {
	return slog.StringValue(f.string)
}
