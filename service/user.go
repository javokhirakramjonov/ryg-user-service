package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"log"
	pbe "ryg-user-service/gen_proto/email_service"
	pbu "ryg-user-service/gen_proto/user_service"
	"ryg-user-service/model"
	"ryg-user-service/rabbit_mq"
)

type UserService struct {
	db                    *gorm.DB
	genericEmailPublisher *rabbit_mq.GenericEmailPublisher
	pbu.UnimplementedUserServiceServer
}

func NewUserService(db *gorm.DB, genericEmailPublisher *rabbit_mq.GenericEmailPublisher) *UserService {
	return &UserService{
		db:                    db,
		genericEmailPublisher: genericEmailPublisher,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *pbu.CreateUserRequest) (*pbu.User, error) {
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		FullName: req.FullName,
		Password: hashedPassword,
		Email:    req.Email,
		Role:     "user",
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}

	resp := &pbu.User{
		Id:       user.ID,
		FullName: user.FullName,
		Email:    user.Email,
		Role:     user.Role,
		IsActive: user.IsActive,
	}

	err = s.genericEmailPublisher.Publish(&pbe.GenericEmail{
		To:      user.Email,
		Subject: "Welcome to RYG",
		Body:    "Welcome to RYG, we are glad to have you on board!",
	})

	if err != nil {
		log.Printf("Failed to publish email: %v", err)
	}

	return resp, nil
}

func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (s *UserService) GetUserById(ctx context.Context, req *pbu.GetUserRequest) (*pbu.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, req.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve user: %v", err)
	}

	return &pbu.User{
		Id:       user.ID,
		FullName: user.FullName,
		Email:    user.Email,
		Role:     user.Role,
		IsActive: user.IsActive,
	}, nil
}

func (s *UserService) GetUserForLogin(ctx context.Context, req *pbu.GetUserForLoginRequest) (*pbu.UserForLogin, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve user: %v", err)
	}

	return &pbu.UserForLogin{
		Id:       user.ID,
		Email:    user.Email,
		Password: user.Password,
		Role:     user.Role,
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *pbu.UpdateUserRequest) (*pbu.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, req.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve user: %v", err)
	}

	user.FullName = req.FullName

	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &pbu.User{
		Id:       user.ID,
		FullName: user.FullName,
		Email:    user.Email,
		Role:     user.Role,
		IsActive: user.IsActive,
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *pbu.DeleteUserRequest) (*emptypb.Empty, error) {
	if err := s.db.WithContext(ctx).Delete(&model.User{}, req.Id).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}
	return &emptypb.Empty{}, nil
}
