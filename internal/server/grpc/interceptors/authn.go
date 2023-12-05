package interceptors

import (
	"context"
	"fmt"
	"github.com/bobgromozeka/yp-diploma2/internal/jwt"
	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

var protectedServices = []string{
	"datakeeper.DataKeeper",
}

type serverStreamWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (ssw *serverStreamWrapper) Context() context.Context {
	return ssw.ctx
}

const UserID = "userID"

func AuthnUnary(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	for _, service := range protectedServices {
		if strings.Contains(info.FullMethod, service) {
			ctx, ctxErr := createCtxWithUserID(ctx)
			if ctxErr != nil {
				return nil, ctxErr
			}

			return handler(ctx, req)
		}
	}

	return handler(ctx, req)
}

func AuthnStream(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	for _, service := range protectedServices {
		if strings.Contains(info.FullMethod, service) {
			ctx, ctxErr := createCtxWithUserID(ss.Context())
			if ctxErr != nil {
				return ctxErr
			}

			return handler(srv, &serverStreamWrapper{
				ServerStream: ss,
				ctx:          ctx,
			})
		}
	}

	return status.Errorf(codes.Unauthenticated, "Unauthenticated")
}

func createCtxWithUserID(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		fmt.Println("No auth header")
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated")
	}
	var jwtString string
	authHeader := md.Get("Authorization")

	if len(authHeader) == 1 && strings.Contains(authHeader[0], "Bearer ") {
		jwtString = strings.Split(authHeader[0], " ")[1]
	} else {
		fmt.Println("Auth header wrong format: ", authHeader)
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	claims, claimsErr := jwt.GetClaimsFromSign(jwtString)
	if claimsErr != nil || claims == nil {
		fmt.Println("Could not parse jwt: ", claimsErr)
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	uID, ok := claims[jwt.UserIDKey]
	if !ok {
		fmt.Println("No user id in jwt: ", claimsErr)
		return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	return context.WithValue(ctx, UserID, uID), nil
}
