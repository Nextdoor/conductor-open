package core

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/Nextdoor/conductor/services/auth"
	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/shared/types"
)

type TestData struct {
	Train       *types.Train
	User        *types.User
	TokenCookie *http.Cookie
}

var token uint64
var robotCreated bool

func setup(t *testing.T) (*mux.Router, *TestData) {
	endpoints := Endpoints()
	conductorServer := NewServer(endpoints)

	dataClient := data.NewClient()

	commit := types.Commit{SHA: "sha1"}
	commits := []*types.Commit{&commit}

	if !robotCreated {
		user, err := dataClient.ReadOrCreateUser("robot", "robot@example.com")
		assert.NoError(t, err)

		err = dataClient.WriteToken("robot", user.Name, user.Email, "", "")
		assert.NoError(t, err)

		robotCreated = true
	}

	user, err := dataClient.ReadOrCreateUser("test_user", "test_email")
	assert.NoError(t, err)

	tokenVal := strconv.FormatUint(token, 10)
	token += 1

	types.CustomizeJobs(types.Delivery, []string{})
	types.CustomizeJobs(types.Verification, []string{})
	types.CustomizeJobs(types.Deploy, []string{})

	err = dataClient.SetMode(types.Manual)
	assert.NoError(t, err)
	err = dataClient.SetOptions(&types.DefaultOptions)
	assert.NoError(t, err)

	err = dataClient.WriteToken(tokenVal, user.Name, user.Email, "", "")
	assert.NoError(t, err)
	train, err := dataClient.CreateTrain("test_train", user, commits)
	assert.NoError(t, err)

	return conductorServer, &TestData{
		Train:       train,
		User:        user,
		TokenCookie: auth.NewCookie(tokenVal),
	}
}

func listen(t *testing.T, server *mux.Router) {
	if err := http.ListenAndServe(":8400", server); err != nil {
		t.Error(err)
	}
}
