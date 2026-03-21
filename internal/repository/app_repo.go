package repository

import (
	"context"

	"github.com/luckysxx/user-platform/internal/dberr"
	"github.com/luckysxx/user-platform/internal/ent"
	"github.com/luckysxx/user-platform/internal/ent/app"
	"github.com/luckysxx/user-platform/internal/ent/user"
	"github.com/luckysxx/user-platform/internal/ent/userappprofile"
)

type AppRepository interface {
	EnsureAppAuthorized(ctx context.Context, userID int64, appCode string) error
}

type appRepository struct {
	client *ent.Client
}

func NewAppRepository(client *ent.Client) AppRepository {
	return &appRepository{client: client}
}

// EnsureAppAuthorized 确保用户已授权应用
func (r *appRepository) EnsureAppAuthorized(ctx context.Context, userID int64, appCode string) error {
	// 使用 Ent 推荐的标准事务闭包写法，保障高并发下的数据强一致性
	return WithTx(ctx, r.client, func(txCtx context.Context) error {
		tx := ent.TxFromContext(txCtx)

		appNode, err := tx.App.Query().Where(app.AppCodeEQ(appCode)).Only(txCtx)
		if err != nil {
			return dberr.ParseDBError(err)
		}

		_, err = tx.UserAppProfile.Query().
			Where(userappprofile.HasUserWith(user.IDEQ(userID))).
			Where(userappprofile.HasAppWith(app.IDEQ(appNode.ID))).
			Only(txCtx)
		if err == nil {
			// 已经授权过了，正常结束（无需再建记录）
			return nil
		}
		if !ent.IsNotFound(err) {
			return dberr.ParseDBError(err) // 如果是其他 DB 查询报错，中止事务
		}

		_, err = tx.UserAppProfile.Create().
			SetUserID(userID).
			SetAppID(appNode.ID).
			Save(txCtx)
		if err != nil {
			if ent.IsConstraintError(err) {
				// 极高并发下的“唯一约束冲突”（两个请求挤过了 SELECT 判断同时尝试 CREATE）
				// 此时忽略报错当做成功即可，因为记录已被另一个协程写进去了
				return nil
			}
			return dberr.ParseDBError(err)
		}
		return nil
	})
}
