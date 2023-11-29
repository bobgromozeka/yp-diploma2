package services

import (
	"context"
	"errors"
	"log"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/user"
	"github.com/bobgromozeka/yp-diploma2/internal/server/storage"
)

const minPasswordLen = 8

var (
	ErrInternalServerError = status.Errorf(codes.Internal, "Internal server error")
)

const JWTSecretKey = "6ecafb3785ddd92172f71ec4821211e124e0527a268cb6445c8fdaa02c6a2628f25bfb17ef95f1a0873ab87f6f559958deccb7c7902514f0164efd99950a670c"

type UserService struct {
	user.UnimplementedUserServer
	us storage.UserStorage
}

func NewUserService(us storage.UserStorage) *UserService {
	return &UserService{
		us: us,
	}
}

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

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.MapClaims{
			"userID": u.ID,
		},
	)

	strToken, signErr := token.SignedString([]byte(JWTSecretKey))
	if signErr != nil {
		log.Default().Println("Error while signing JWT key: ", signErr)
		return nil, ErrInternalServerError
	}

	return &user.SignInResponse{
		Token: &strToken,
	}, nil
}
