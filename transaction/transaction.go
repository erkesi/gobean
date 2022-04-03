package transaction

import (
	"context"
	"database/sql"
)

type TxMgr interface {
	OpenMainTx(ctx context.Context, news ...bool) (context.Context, int64)
	CloseMainTx(ctx context.Context, txId int64, err error)
	MainDb(ctx context.Context) sql.DB
}
