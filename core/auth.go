package core

import (
	"context"
	"fmt"
	"net/http"

	uuid "github.com/satori/go.uuid"

	"github.com/Nextdoor/conductor/services/auth"
	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/settings"
	"github.com/Nextdoor/conductor/shared/types"
)

func newAuthMiddleware() authMiddleware {
	return authMiddleware{}
}

type authMiddleware struct{}

const AdminPermissionMessage = "Only admins can call this endpoint."

func (_ authMiddleware) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler.(endpoint).needsAuth {
			cookie, err := r.Cookie(auth.GetCookieName())
			// Note: Only possible error is ErrNoCookie.
			if err == http.ErrNoCookie {
				errorResponse("Unauthorized", http.StatusUnauthorized).Write(w, r)
				return
			}
			token := cookie.Value

			dataClient := data.NewClient()
			user, err := dataClient.UserByToken(token)
			if err != nil {
				logger.Error("Error getting user by token (%s): %v", token, err)
				errorResponse("Unauthorized", http.StatusUnauthorized).Write(w, r)
				return
			}
			user.IsAdmin = settings.IsAdminUser(user.Email)

			// Check admin access restrictions.
			if handler.(endpoint).needsAdmin && !user.IsAdmin {
				errorResponse(
					fmt.Sprintf("%s You are logged in as %s.",
						AdminPermissionMessage, user.Name),
					http.StatusForbidden).Write(w, r)
				return
			}

			handler.ServeHTTP(w, r.WithContext(
				context.WithValue(r.Context(), "user", user)))
		} else {
			handler.ServeHTTP(w, r)
		}
	})
}

func authEndpoints() []endpoint {
	return []endpoint{
		newOpenEp("/api/auth/info", get, authInfo),
		newOpenEp("/api/auth/login", get, authLogin),
		newEp("/api/auth/logout", post, authLogout),
	}
}

// Provides auth info.
// This endpoint is currently unused, but we might want it in the future (cli?).
func authInfo(_ *http.Request) response {
	authService := auth.GetService()
	authURL := authService.AuthURL(settings.GetHostname())
	authProvider := authService.AuthProvider()
	return dataResponse(struct {
		URL      string `json:"url"`
		Provider string `json:"provider"`
	}{
		authURL,
		authProvider,
	})
}

func authLogin(r *http.Request) response {
	authService := auth.GetService()
	dataClient := data.NewClient()
	code, ok := r.URL.Query()["code"]
	if !ok {
		return errorResponse("'code' must be included in post form.", http.StatusBadRequest)
	}
	if len(code) != 1 {
		return errorResponse(fmt.Sprintf("'code' in post form had %d elements; 1 expected.", len(code)),
			http.StatusBadRequest)
	}
	name, email, avatar, codeToken, err := authService.Login(code[0])
	if err != nil {
		return errorResponse(err.Error(), http.StatusInternalServerError)
	}
	if name == "" || email == "" {
		return errorResponse(
			fmt.Sprintf("Name, email, and avatar must be set, were %s, %s, and %s respectively.",
				name, email, avatar), http.StatusInternalServerError)
	}
	token := uuid.NewV4().String() // TODO: Read from env for robot user.
	err = dataClient.WriteToken(token, name, email, avatar, codeToken)
	if err != nil {
		return errorResponse(err.Error(), http.StatusInternalServerError)
	}

	return loginResponse(token)
}

func authLogout(r *http.Request) response {
	dataClient := data.NewClient()
	authedUser := r.Context().Value("user").(*types.User)
	err := dataClient.RevokeToken(authedUser.Token, authedUser.Email)
	if err != nil {
		return errorResponse(err.Error(), http.StatusInternalServerError)
	}
	return logoutResponse()
}

func loginResponse(token string) response {
	return response{
		Code:         http.StatusFound,
		Cookie:       auth.NewCookie(token),
		RedirectPath: "/",
	}
}

func logoutResponse() response {
	return response{
		Code:   http.StatusOK,
		Cookie: auth.EmptyCookie(),
	}
}
