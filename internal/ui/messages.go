package ui

import (
	"github.com/msteinert/pam/v2"
	minidm "github.com/r4ppz/minidm/internal"
)

type CredentialSubmitMsg struct {
	Username string
	Password string
}

type AuthResultMsg struct {
	Err      error
	Username string
	PAMTx    *pam.Transaction
	Session  minidm.Session
}
