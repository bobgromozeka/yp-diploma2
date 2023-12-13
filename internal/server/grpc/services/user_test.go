package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/user"
	"github.com/bobgromozeka/yp-diploma2/internal/server/storage"
	"github.com/bobgromozeka/yp-diploma2/pkg/testutils"
)

func TestUserService_SignUp_ShortPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := NewUserService(storage.NewMockUserStorage(ctrl))

	resp, err := s.SignUp(context.Background(), &user.SignUpRequest{
		Login:    "login",
		Password: "pass",
	})

	assert.Nil(t, err)
	assert.Equal(t, []*user.FieldError{
		{
			Name:  "password",
			Error: "Password should be of length 8 or more",
		},
	}, resp.Errors)
}

func TestUserService_SignUp_LoginExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := storage.NewMockUserStorage(ctrl)
	us.
		EXPECT().
		CreateUser(testutils.MatchContext(), gomock.Eq("login"), gomock.AssignableToTypeOf([]byte{})).
		Return(storage.ErrLoginAlreadyExists)

	s := NewUserService(us)

	resp, err := s.SignUp(context.Background(), &user.SignUpRequest{
		Login:    "login",
		Password: "password",
	})

	assert.Nil(t, err)
	assert.Equal(t, []*user.FieldError{
		{
			Name:  "login",
			Error: "Login already exists",
		},
	}, resp.Errors)
}

func TestUserService_SignUp_InternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := storage.NewMockUserStorage(ctrl)
	us.
		EXPECT().
		CreateUser(testutils.MatchContext(), gomock.Eq("login"), gomock.AssignableToTypeOf([]byte{})).
		Return(errors.New("random error"))

	s := NewUserService(us)

	resp, err := s.SignUp(context.Background(), &user.SignUpRequest{
		Login:    "login",
		Password: "password",
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestUserService_SignUp_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := storage.NewMockUserStorage(ctrl)
	us.
		EXPECT().
		CreateUser(testutils.MatchContext(), gomock.Eq("login"), gomock.AssignableToTypeOf([]byte{})).
		Return(nil)

	s := NewUserService(us)

	resp, err := s.SignUp(context.Background(), &user.SignUpRequest{
		Login:    "login",
		Password: "password",
	})

	assert.Nil(t, err)
	assert.Equal(t, &user.SignUpResponse{
		Success: true,
	}, resp)
}

func TestUserService_SignIn_InternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := storage.NewMockUserStorage(ctrl)
	us.
		EXPECT().
		GetUser(testutils.MatchContext(), gomock.Eq("login")).
		Return(nil, errors.New("random error"))

	s := NewUserService(us)

	resp, err := s.SignIn(context.Background(), &user.SignInRequest{
		Login:    "login",
		Password: "password",
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestUserService_SignIn_UnknownUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := storage.NewMockUserStorage(ctrl)
	us.
		EXPECT().
		GetUser(testutils.MatchContext(), gomock.Eq("login")).
		Return(nil, nil)

	s := NewUserService(us)

	resp, err := s.SignIn(context.Background(), &user.SignInRequest{
		Login:    "login",
		Password: "password",
	})

	assert.Nil(t, err)
	assert.Equal(t, &user.SignInResponse{
		Token: nil,
		Error: "Wrong login or password",
	}, resp)
}

func TestUserService_SignIn_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := storage.NewMockUserStorage(ctrl)
	us.
		EXPECT().
		GetUser(testutils.MatchContext(), gomock.Eq("login")).
		Return(&storage.User{
			ID:       1,
			Login:    "login",
			Password: "passwordpassword",
		}, nil)

	s := NewUserService(us)

	resp, err := s.SignIn(context.Background(), &user.SignInRequest{
		Login:    "login",
		Password: "password",
	})

	assert.Nil(t, err)
	assert.Equal(t, &user.SignInResponse{
		Token: nil,
		Error: "Wrong login or password",
	}, resp)
}

func TestUserService_SignIn_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	us := storage.NewMockUserStorage(ctrl)
	us.
		EXPECT().
		GetUser(testutils.MatchContext(), gomock.Eq("login")).
		Return(&storage.User{
			ID:       1,
			Login:    "login",
			Password: "$2a$10$Kx7XUlICoBuqilJ9hoAUd.bhpjzbtXICijmd9hf6r9sT2NkzI3ryK",
		}, nil)

	s := NewUserService(us)

	resp, err := s.SignIn(context.Background(), &user.SignInRequest{
		Login:    "login",
		Password: "password",
	})

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOjF9.fXMEMQ5FGglGA64ijzTPDf-94QSY_Bv1KLcbkHmPO1s"
	assert.Nil(t, err)
	assert.Equal(t, &user.SignInResponse{
		Token: &token,
	}, resp)
}
