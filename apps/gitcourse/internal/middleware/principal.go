package middleware

import (
	"context"
	"net/http"

	hubrelay "github.com/fastygo/hubrelay-sdk"
)

type principalKey struct{}

var defaultPrincipal = hubrelay.Principal{
	ID:      "gitcourse",
	Display: "gitcourse",
	Roles:   []string{"operator"},
}

func WithPrincipal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), principalKey{}, defaultPrincipal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func PrincipalFromContext(ctx context.Context) hubrelay.Principal {
	if principal, ok := ctx.Value(principalKey{}).(hubrelay.Principal); ok {
		return principal
	}
	return defaultPrincipal
}
