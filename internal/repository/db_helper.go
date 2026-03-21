package repository

import (
	"context"
	"fmt"

	"github.com/luckysxx/user-platform/internal/ent"
)

// WithTx 封装好的事务大管家函数 (Transaction Wrapper)
func WithTx(ctx context.Context, client *ent.Client, fn func(ctx context.Context) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("starting a transaction: %w", err)
	}

	// 捕获 panic 并确保安全回滚
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	txCtx := ent.NewTxContext(ctx, tx)

	// 执行业务逻辑
	if err := fn(txCtx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
		}
		return err
	}

	// 顺利提交
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// 事务管理器接口 (TransactionManager)
type TransactionManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type transactionManager struct {
	client *ent.Client
}

// NewTransactionManager 构造函数，在 cmd/main.go 中注入
func NewTransactionManager(client *ent.Client) TransactionManager {
	return &transactionManager{client: client}
}

// 实现接口：内部偷偷调用了上面写好的带魔法包裹的 WithTx
func (tm *transactionManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return WithTx(ctx, tm.client, fn)
}
