package services

import (
	"context"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/user"
	"github.com/bobgromozeka/yp-diploma2/internal/jwt"
	"github.com/bobgromozeka/yp-diploma2/internal/server/storage"
)

const minPasswordLen = 8

var (
	ErrInternalServerError = status.Errorf(codes.Internal, "Internal server error")
)

// UserService implementation of gRPC user service
type UserService struct {
	user.UnimplementedUserServer
	us storage.UserStorage
}

func NewUserService(us storage.UserStorage) *UserService {
	return &UserService{
		us: us,
	}
}

// SignUp Creates new user. Returns error if password is too short or if login already exists.
func (s *UserService) SignUp(ctx context.Context, req *user.SignUpRequest) (*user.SignUpResponse, error) {
	if len(req.Password) < minPasswordLen {
		return &user.SignUpResponse{
			Success: false,
			Errors: []*user.FieldError{
				{
					Name:  "password",
					Error: "Password should be of length 8 or more",
				},
			},
		}, nil
	}

	encryptedPassword, encryptError := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if encryptError != nil {
		log.Default().Println(encryptError)
		return nil, ErrInternalServerError
	}

	err := s.us.CreateUser(ctx, req.Login, encryptedPassword)
	if err != nil {
		if errors.Is(err, storage.ErrLoginAlreadyExists) {
			return &user.SignUpResponse{
				Success: false,
				Errors: []*user.FieldError{
					{
						Name:  "login",
						Error: "Login already exists",
					},
				},
			}, nil
		}
		return nil, ErrInternalServerError
	}

	return &user.SignUpResponse{
		Success: true,
	}, nil
}

// SignIn authenticates user. Checks login and password and creates new JWT if no errors occurred.
func (s *UserService) SignIn(ctx context.Context, req *user.SignInRequest) (*user.SignInResponse, error) {
	u, uErr := s.us.GetUser(ctx, req.Login)
	if uErr != nil {
		return nil, ErrInternalServerError
	}

	if u == nil {
		return &user.SignInResponse{
			Token: nil,
			Error: "Wrong login or password",
		}, nil
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); compareErr != nil {
		return &user.SignInResponse{
			Token: nil,
			Error: "Wrong login or password",
		}, nil
	}

	strToken, signErr := jwt.NewJWTWithUserID(u.ID)
	if signErr != nil {
		log.Default().Println("Error while signing JWT key: ", signErr)
		return nil, ErrInternalServerError
	}

	return &user.SignInResponse{
		Token: &strToken,
	}, nil
}
