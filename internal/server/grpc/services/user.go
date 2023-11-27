package services

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/user"
)

const minPasswordLen = 8

var (
	ErrInternalServerError = status.Errorf(codes.Internal, "Internal server error")
	ErrInvalidArgument     = status.Errorf(codes.InvalidArgument, "arguments error")
)

const JWTSecretKey = "6ecafb3785ddd92172f71ec4821211e124e0527a268cb6445c8fdaa02c6a2628f25bfb17ef95f1a0873ab87f6f559958deccb7c7902514f0164efd99950a670c"

type UserService struct {
	user.UnimplementedUserServer
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		db: db,
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

	_, err := s.db.ExecContext(
		ctx, "insert into users (login, password) values ($1, $2)", req.Login, encryptedPassword,
	)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
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
		}
		return nil, ErrInternalServerError
	}

	return &user.SignUpResponse{
		Success: true,
	}, nil
}

func (s *UserService) SignIn(ctx context.Context, req *user.SignInRequest) (*user.SignInResponse, error) {
	u := s.db.QueryRowContext(ctx, "select id, password from users where login = $1", req.Login)

	var pass string
	var id int64
	if scanErr := u.Scan(&id, &pass); scanErr != nil {
		if errors.Is(scanErr, sql.ErrNoRows) {
			return &user.SignInResponse{
				Token: nil,
				Error: "Wrong login or password",
			}, nil
		}
		log.Default().Println("[SignIn] query user error: ", scanErr)
		return nil, ErrInternalServerError
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(pass), []byte(req.Password)); compareErr != nil {
		return &user.SignInResponse{
			Token: nil,
			Error: "Wrong login or password",
		}, nil
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.MapClaims{
			"userID": id,
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
