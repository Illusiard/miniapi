package caps

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Store interface {
	Ping(ctx context.Context) error
	Exec(ctx context.Context, sql string, args ...any) error
	RunInTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}
