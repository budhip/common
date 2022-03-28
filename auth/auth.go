package auth

import (
	"context"
	"net/http"
	"strings"

	cctx "github.com/budhip/common/context"
	"google.golang.org/grpc/metadata"
)

const (
	bearer        string = "bearer"
	authorization string = "authorization"
	jwtpayload    string = "jwtpayload"
	cID			  string = "cID"
)

type UserInfo struct {
	ID     uint64 `json:"user_id,omitempty"`
	Name   string `json:"name,omitempty"`
	Mobile string `json:"phone_number,omitempty"`
	Email  string `json:"email,omitempty"`
}

func extractTokenFromAuthHeader(auth string) (string, bool) {
	authHeaderParts := strings.Split(auth, " ")
	if len(authHeaderParts) != 2 || !strings.EqualFold(authHeaderParts[0], bearer) {
		return "", false
	}

	return authHeaderParts[1], true
}

func extractPayloadFromToken(token string) (string, bool) {
	tokenParts := strings.Split(token, ".")
	if len(tokenParts) != 3 {
		return "", false
	}

	return tokenParts[1], true
}

func withUserInfo(ctx context.Context, payload string) context.Context {
	//var userInfo UserInfo
	//if claims, err := base64.RawURLEncoding.DecodeString(payload); err != nil {
	//	log.Printf("error while decoding jwt payload: %v", err)
	//
	//	return ctx
	//} else if err := json.Unmarshal(claims, &userInfo); err != nil {
	//	log.Printf("error while unmarshalling jwt payload: %v", err)
	//
	//	return ctx
	//}

	return context.WithValue(ctx, cctx.CtxCID, payload)
}

func WithUserInfoContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	//var payload string
	//payloadHeader, ok := md[jwtpayload]
	//if ok {
	//	payload = payloadHeader[0]
	//} else {
	//	authHeader, ok := md[authorization]
	//	if !ok {
	//		return ctx
	//	}
	//
	//	token, ok := extractTokenFromAuthHeader(authHeader[0])
	//	if ok {
	//		payload, ok = extractPayloadFromToken(token)
	//		if !ok {
	//			return ctx
	//		}
	//	}
	//}

	var payload string
	payloadHeader, ok := md[cID]
	if ok {
		payload = payloadHeader[0]
	} else {
		return ctx
	}

	return withUserInfo(ctx, payload)
}

func WithUserInfoRequestContext(req *http.Request) *http.Request {
	ctx := req.Context()

	var payload string
	payloadHeader := req.Header.Get(jwtpayload)
	if len(payloadHeader) > 0 {
		payload = payloadHeader
	} else {
		authHeader := req.Header.Get(authorization)
		if len(authHeader) > 0 {
			token, ok := extractTokenFromAuthHeader(authHeader)
			if ok {
				payload, ok = extractPayloadFromToken(token)
				if !ok {
					return req
				}
			} else {
				return req
			}
		} else {
			return req
		}
	}

	ctx = withUserInfo(ctx, payload)
	return req.WithContext(ctx)
}