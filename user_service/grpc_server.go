package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/akmmp241/topupstore-microservice/shared"
	upb "github.com/akmmp241/topupstore-microservice/user-proto/v1"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcServer struct {
	ListenAddr  string
	DB          *sql.DB
	Server      *grpc.Server
	NetListener net.Listener
	upb.UnimplementedUserServiceServer
}

func NewGrpcServer(listenAddr string, DB *sql.DB) *GrpcServer {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		slog.Error("Error occurred while creating listener", "err", err)
		panic(err)
	}

	return &GrpcServer{
		ListenAddr:  listenAddr,
		DB:          DB,
		Server:      grpc.NewServer(),
		NetListener: listener,
	}
}

func (s *GrpcServer) Run() {
	upb.RegisterUserServiceServer(s.Server, s)

	slog.Info("Starting gRPC server", "addr", s.ListenAddr)

	if err := s.Server.Serve(s.NetListener); err != nil {
		slog.Error("Error occurred while serving gRPC server", "err", err)
		panic(err)
	}
}

func (s *GrpcServer) CreateUser(ctx context.Context, req *upb.CreateUserReq) (*upb.CreateUserRes, error) {
	password, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Error occurred while hashing password", "err", err)
		return nil, err
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer shared.CommitOrRollback(tx, err)

	result, err := tx.ExecContext(ctx, "INSERT INTO users (id, name, email, password,phone_number, email_verification_token) VALUES (NULL, ?, ?, ?, ?, ?)",
		req.Name, req.Email, string(password), req.PhoneNumber, req.EmailVerificationToken)

	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			slog.Error("MySQL error occurred", "code", mysqlErr.Number, "message", mysqlErr.Message)
			// check for duplicate entry error
			if mysqlErr.Number == 1062 {
				errMsg := ""
				// parse the error message to find which constraint was violated
				if strings.Contains(mysqlErr.Message, "email") {
					errMsg = "email already exists"
				} else if strings.Contains(mysqlErr.Message, "phone_number") {
					errMsg = "phone number already exists"
				}
				return nil, status.Error(codes.AlreadyExists, "Duplicate entry: "+errMsg)
			}
		}
		slog.Info("Error occurred while inserting user", "err", err)
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		slog.Info("No rows affected while inserting user", "err", err)
		return nil, err
	}

	return &upb.CreateUserRes{Msg: "Success create user"}, nil
}

func (s *GrpcServer) GetUserById(ctx context.Context, req *upb.GetUserByIdReq) (*upb.GetUserRes, error) {
	user, err := s.getUser(ctx, req.Id, "id")
	if err != nil {
		return nil, err
	}

	getUserRes := upb.GetUserRes{
		User: user,
	}

	return &getUserRes, nil
}

func (s *GrpcServer) GetUserByEmail(ctx context.Context, req *upb.GetUserByEmailReq) (*upb.GetUserRes, error) {
	user, err := s.getUser(ctx, req.Email, "email")
	if err != nil {
		return nil, err
	}

	getUserRes := upb.GetUserRes{
		User: user,
	}

	return &getUserRes, nil
}

func (s *GrpcServer) ResetPasswordByEmail(ctx context.Context, req *upb.ResetPasswordByEmailReq) (*upb.ResetPasswordRes, error) {
	password, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Error occurred while hashing password", "err", err)
		return nil, err
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer shared.CommitOrRollback(tx, err)

	query := "UPDATE users SET password = ?, updated_at = CURRENT_TIMESTAMP WHERE email = ?"

	result, err := tx.ExecContext(ctx, query, string(password), req.GetEmail())
	if err != nil {
		slog.Info("Internal server error", "err", err)
		return nil, err
	}

	// checks update status
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		slog.Info("No rows affected while updating user", "err", err)
		return nil, status.Error(codes.NotFound, "User not found")
	}

	return &upb.ResetPasswordRes{Msg: "successfully reset password"}, nil
}

func (s *GrpcServer) VerifyEmail(ctx context.Context, req *upb.VerifyEmailReq) (*upb.VerifyEmailRes, error) {
	emailVerifiedAt := time.Now()

	query := "UPDATE users SET email_verification_token = NULL, email_verified_at = ? WHERE email_verification_token = ? AND email_verified_at IS NULL"
	result, err := s.DB.ExecContext(ctx, query, emailVerifiedAt, req.GetEmailVerificationToken())
	if err != nil {
		slog.Error("Error occurred while updating user", "err", err)
		return nil, err
	}

	// checks update status
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		slog.Info("No rows affected while updating user", "err", err)
		return nil, status.Error(codes.NotFound, "User not found")
	}

	return &upb.VerifyEmailRes{Msg: "successfully verified email"}, nil
}

func (s *GrpcServer) getUser(ctx context.Context, target string, column string) (*upb.User, error) {
	query := fmt.Sprintf("SELECT id, name, email, password, phone_number, email_verified_at, created_at, updated_at FROM users WHERE %s = ?", column)

	rows, err := s.DB.QueryContext(ctx, query, target)
	if err != nil {
		slog.Error("Error occurred while querying user", "err", err)
		return nil, err
	}
	defer rows.Close()

	var user upb.User
	if !rows.Next() {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	var emailVerifiedAt sql.NullTime
	var createdAt time.Time
	var updatedAt time.Time
	err = rows.Scan(&user.Id, &user.Name, &user.Email, &user.Password, &user.PhoneNumber, &emailVerifiedAt, &createdAt, &updatedAt)
	if err != nil {
		slog.Error("Error occurred while scanning user", "err", err)
		return nil, err
	}

	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = timestamppb.New(emailVerifiedAt.Time)
	}
	user.CreatedAt = timestamppb.New(createdAt)
	user.UpdatedAt = timestamppb.New(updatedAt)

	return &user, nil
}
