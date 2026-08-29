package ui

import minidm "github.com/r4ppz/minidm/internal"

type AuthResultMsg struct {
	Err error
}

type CredentialSubmitMsg struct {
	Username string
	Password string
	Session  minidm.Session
}
