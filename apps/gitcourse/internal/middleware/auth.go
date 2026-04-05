package middleware

import coreauth "github.com/fastygo/hubcore/auth"

type SessionAuth = coreauth.SessionAuth

var NewSessionAuth = coreauth.NewSessionAuth

const sessionCookieName = "hubrelay_session"
