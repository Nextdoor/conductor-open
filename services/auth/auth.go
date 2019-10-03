/* Handles authenticating requests and performing oauth. */
package auth

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Nextdoor/conductor/shared/flags"
	"github.com/Nextdoor/conductor/shared/logger"
)

var (
	implementationFlag = flags.EnvString("AUTH_IMPL", "fake")

	// This cookie name has to match the cookie name clients expect.
	authCookieName = flags.EnvString("AUTH_COOKIE_NAME", "conductor-auth")
)

type Service interface {
	AuthProvider() string
	AuthURL(hostname string) string
	Login(code string) (name, email, avatar, codeToken string, err error)
}

type auth struct{}

func GetCookieName() string {
	return authCookieName
}

func NewCookie(token string) *http.Cookie {
	return &http.Cookie{Name: GetCookieName(), Value: token, Path: "/"}
}

func EmptyCookie() *http.Cookie {
	return &http.Cookie{Name: GetCookieName(), Value: "", Path: "/", Expires: time.Time{}}
}

var (
	service Service
	getOnce sync.Once
)

func GetService() Service {
	getOnce.Do(func() {
		service = newService()
	})
	return service
}

func newService() Service {
	logger.Info("Using %s implementation for Auth service", implementationFlag)
	var service Service
	switch implementationFlag {
	case "fake":
		service = newFake()
	case "github":
		service = newGithubAuth()
	default:
		panic(fmt.Errorf("Unknown Auth Implementation: %s", implementationFlag))
	}
	return service
}

type fake struct{}

func newFake() *fake {
	return &fake{}
}

func redirectEndpoint(hostname string) string {
	return fmt.Sprintf("%s/api/auth/done", hostname)
}

func (a *fake) AuthProvider() string {
	return ""
}

func (a *fake) AuthURL(hostname string) string {
	return ""
}

func (a *fake) Login(code string) (name, email, avatar, codeToken string, err error) {
	// If a developer doesn't choose to do github setup in envfile,
	// this should still allow going past the login page, without fetching
	// github profile details and avatar
	return "dev", "dev@conductor.com", "", "", nil
}
