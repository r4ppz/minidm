package ui

import (
	"github.com/msteinert/pam/v2"
	"github.com/r4ppz/minidm/internal/session"
)

type CredentialSubmitMsg struct {
	Username string
	Password string
}

type AuthResultMsg struct {
	Err      error
	Username string
	PAMTx    *pam.Transaction
	Session  session.Session
}
