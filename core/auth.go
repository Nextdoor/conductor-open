package core

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/shared/datadog"
	"github.com/Nextdoor/conductor/shared/flags"
	"github.com/Nextdoor/conductor/shared/logger"
	"github.com/Nextdoor/conductor/shared/settings"
	"github.com/Nextdoor/conductor/shared/types"
)

var (
	// The headers to trust for auth.
	// This header's value should be the authed user's name. It must be set.
	// If this header is not set, authentication will be denied.
	authHeaderName = flags.EnvString("AUTH_HEADER_NAME", "X-Conductor-User")
	// This header's value should be the authed user's email. It must be set.
	// If this header is not set, authentication will be denied.
	// This is checked against VIEWER_EMAILS to admit viewers.
	// This is checked against USER_EMAILS to admit users.
	// This is checked against ADMIN_EMAILS to admit admins.
	authHeaderEmail = flags.EnvString("AUTH_HEADER_EMAIL", "X-Conductor-Email")
	// This optional header's value should be a comma-seperated string of the authed user's groups.
	// This is checked against VIEWER_EMAILS to admit viewers.
	// This is checked against USER_EMAILS to admit users.
	// This is checked against ADMIN_EMAILS to admit admins.
	authHeaderGroups = flags.EnvString("AUTH_HEADER_GROUPS", "X-Conductor-Groups")

	// Debug auth values - Only use this for dev!
	devAuthNameOverride   = flags.EnvString("DEV_AUTH_NAME_VALUE", "")
	devAuthEmailOverride  = flags.EnvString("DEV_AUTH_EMAIL_VALUE", "")
	devAuthGroupsOverride = flags.EnvString("DEV_AUTH_GROUPS_VALUE", "")
)

func newAuthMiddleware() authMiddleware {
	if devAuthNameOverride != "" {
		datadog.Info("Auth Name override set: %s", devAuthNameOverride)
	}

	if devAuthEmailOverride != "" {
		datadog.Info("Auth Email override set: %s", devAuthEmailOverride)
	}

	if devAuthGroupsOverride != "" {
		datadog.Info("Auth Groups override set: %s", devAuthGroupsOverride)
	}

	return authMiddleware{}
}

type authMiddleware struct{}

const AdminPermissionMessage = "Only Conductor admins can use this."
const UserPermissionMessage = "Only Conductor users can use this."
const ViewerPermissionMessage = "Only Conductor viewers can see this."

func (_ authMiddleware) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !handler.(endpoint).NeedsAuth() {
			handler.ServeHTTP(w, r)
		} else {
			name := r.Header.Get(authHeaderName)
			email := r.Header.Get(authHeaderEmail)
			groups := r.Header.Get(authHeaderGroups)

			fmt.Println("Headers", r.Header)

			if name == "" {
				if devAuthNameOverride != "" {
					name = devAuthNameOverride
				}
			}

			if email == "" {
				if devAuthEmailOverride != "" {
					email = devAuthEmailOverride
				}
			}

			if groups == "" {
				if devAuthGroupsOverride != "" {
					groups = devAuthGroupsOverride
				}
			}

			if name == "" || email == "" {
				errorResponse("Unauthorized", http.StatusUnauthorized).Write(w, r)
				return
			}

			var user *types.User
			var err error

			dataClient := data.NewClient()
			user, err = dataClient.ReadOrCreateUser(name, email)
			if err != nil {
				logger.Error("Error getting user by email (%s): %v", email, err)
				errorResponse("Unauthorized", http.StatusUnauthorized).Write(w, r)
				return
			}

			user.IsViewer = settings.IsViewer(user.Email, groups)
			user.IsUser = settings.IsUser(user.Email, groups)
			user.IsAdmin = settings.IsAdmin(user.Email, groups)

			if handler.(endpoint).needsAdmin && !user.IsAdmin {
				// Check admin access restrictions.
				errorResponse(
					fmt.Sprintf("%s You are logged in as %s (%s).",
						AdminPermissionMessage, user.Name, user.Email),
					http.StatusForbidden).Write(w, r)
			} else if handler.(endpoint).needsUser && !user.IsUser && !user.IsAdmin {
				// Check user access restrictions.
				errorResponse(
					fmt.Sprintf("%s You are logged in as %s (%s).",
						UserPermissionMessage, user.Name, user.Email),
					http.StatusForbidden).Write(w, r)
			} else if handler.(endpoint).needsViewer && !user.IsViewer && !user.IsUser && !user.IsAdmin {
				// Check viewer access restrictions.
				errorResponse(
					fmt.Sprintf("%s You are logged in as %s (%s).",
						ViewerPermissionMessage, user.Name, user.Email),
					http.StatusForbidden).Write(w, r)
			} else {
				handler.ServeHTTP(w, r.WithContext(
					context.WithValue(r.Context(), "user", user)))
			}
		}
	})
}
