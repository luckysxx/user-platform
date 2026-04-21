package appcontainer

import (
	"github.com/redis/go-redis/v9"

	"github.com/luckysxx/common/ratelimiter"
	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/ent"
	"github.com/luckysxx/user-platform/internal/platform/config"
	"github.com/luckysxx/user-platform/internal/platform/smsauth"
	entaccountstore "github.com/luckysxx/user-platform/internal/store/entstore/account"
	entapplicationstore "github.com/luckysxx/user-platform/internal/store/entstore/application"
	entinfrastore "github.com/luckysxx/user-platform/internal/store/entstore/infra"
	entsessionstore "github.com/luckysxx/user-platform/internal/store/entstore/session"
	redisphonecodestore "github.com/luckysxx/user-platform/internal/store/redisstore/phonecode"
	redissessionstore "github.com/luckysxx/user-platform/internal/store/redisstore/session"
	"go.uber.org/zap"
)

func buildStores(entClient *ent.Client, redisClient *redis.Client) storeSet {
	return storeSet{
		userRepo:       entaccountstore.NewUserStore(entClient),
		identityRepo:   entaccountstore.NewUserIdentityStore(entClient),
		profileRepo:    entaccountstore.NewProfileStore(entClient),
		outboxStore:    entinfrastore.NewEventOutboxStore(entClient),
		tm:             entinfrastore.NewTransactionManager(entClient),
		sessionRepo:    redissessionstore.NewSessionStore(redisClient),
		authzRepo:      entapplicationstore.NewUserAppAuthorizationStore(entClient),
		ssoSessionRepo: entsessionstore.NewSsoSessionStore(entClient),
		appSessionRepo: entsessionstore.NewAppSessionStore(entClient),
		phoneCodeRepo:  redisphonecodestore.NewPhoneCodeStore(redisClient),
	}
}

func buildSupport(cfg *config.Config, redisClient *redis.Client, log *zap.Logger) supportSet {
	return supportSet{
		jwtManager:    auth.NewJWTManager(cfg.JWT.Secret),
		limiter:       ratelimiter.NewFixedWindowLimiter(redisClient, log),
		smsAuthSender: smsauth.NewAliyunSender(cfg.SMSAuth, log),
	}
}
