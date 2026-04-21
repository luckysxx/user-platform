package appcontainer

import (
	"github.com/luckysxx/common/ratelimiter"
	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/platform/smsauth"
	accountrepo "github.com/luckysxx/user-platform/internal/repository/account"
	applicationrepo "github.com/luckysxx/user-platform/internal/repository/application"
	infrarepo "github.com/luckysxx/user-platform/internal/repository/infra"
	sessionrepo "github.com/luckysxx/user-platform/internal/repository/session"
	accountservice "github.com/luckysxx/user-platform/internal/service/account"
	authservice "github.com/luckysxx/user-platform/internal/service/auth"
	"github.com/luckysxx/user-platform/internal/transport/http/server/handler"
)

// Container 承载应用运行所需的核心服务与传输层适配器。
type Container struct {
	UserService    accountservice.UserService
	ProfileService accountservice.ProfileService
	AuthService    authservice.AuthService
	UserHandler    *handler.UserHandler
	JWTManager     *auth.JWTManager
}

type storeSet struct {
	userRepo       accountrepo.UserRepository
	identityRepo   accountrepo.UserIdentityRepository
	profileRepo    accountrepo.ProfileRepository
	outboxStore    infrarepo.EventOutboxWriter
	tm             infrarepo.TransactionManager
	sessionRepo    sessionrepo.SessionRepository
	authzRepo      applicationrepo.UserAppAuthorizationRepository
	ssoSessionRepo sessionrepo.SsoSessionRepository
	appSessionRepo sessionrepo.AppSessionRepository
	phoneCodeRepo  sessionrepo.PhoneCodeRepository
}

type supportSet struct {
	jwtManager    *auth.JWTManager
	limiter       ratelimiter.Limiter
	smsAuthSender smsauth.Sender
}

type serviceSet struct {
	userService    accountservice.UserService
	profileService accountservice.ProfileService
	authService    authservice.AuthService
}
