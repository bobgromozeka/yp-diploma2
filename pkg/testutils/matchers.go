package testutils

import (
	"context"
	"reflect"

	"go.uber.org/mock/gomock"
)

func MatchContext() gomock.Matcher {
	ctx := reflect.TypeOf((*context.Context)(nil)).Elem()
	return gomock.AssignableToTypeOf(ctx)
}
