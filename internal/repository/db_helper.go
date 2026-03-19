package repository

import (
	"context"
	"fmt"

	"github.com/luckysxx/user-platform/internal/ent"
)

// withTx 封装好的事务大管家函数 (Transaction Wrapper)
func withTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
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

	// 执行业务逻辑
	if err := fn(tx); err != nil {
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
