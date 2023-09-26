package Todo

import "time"

const (
	UserTokenValidity        = time.Hour * 24
	UserRefreshTokenValidity = time.Hour * 144
)

const (
	StatusDatabaseCommandOK      = 0
	StatusDatabaseCommandError   = 1
	StatusDatabaseSelectNotFound = 2
)

const (
	StatusGetUserIdSuccess = 0
	StatusUserNotLogin     = 1
	StatusPhraseIdError    = 2
)
