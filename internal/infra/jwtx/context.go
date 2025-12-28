package jwtx

import "context"

type claimsContextKey string

var ctxClaimsKey claimsContextKey = "auth_claims"

// InjectClaims adds token claims into context so other layers can retrieve user info
// รับ Context เดิม จาก Fiber
func InjectClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, ctxClaimsKey, c)
}

// FromContext retrieves claims stored inside context.
// If no claims found, return (nil,false)
func FormContext(ctx context.Context) (*Claims, bool) {
	v := ctx.Value(ctxClaimsKey)
	if v == nil {
		return nil, false
	}
	claims, ok := v.(*Claims)
	return claims, ok
}
