package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"ryg-user-service/db"
	pb "ryg-user-service/gen_proto/user_service"
	"ryg-user-service/model"
)

type UserService struct {
	db *gorm.DB
	pb.UnimplementedUserServiceServer
}

func NewUserService() *UserService {
	return &UserService{
		db: db.DB,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
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

	resp := &pb.User{
		Id:       user.ID,
		FullName: user.FullName,
		Email:    user.Email,
		Role:     user.Role,
		IsActive: user.IsActive,
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

func (s *UserService) GetUserById(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, req.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve user: %v", err)
	}

	return &pb.User{
		Id:       user.ID,
		FullName: user.FullName,
		Email:    user.Email,
		Role:     user.Role,
		IsActive: user.IsActive,
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
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

	return &pb.User{
		Id:       user.ID,
		FullName: user.FullName,
		Email:    user.Email,
		Role:     user.Role,
		IsActive: user.IsActive,
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	if err := s.db.WithContext(ctx).Delete(&model.User{}, req.Id).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}
	return &emptypb.Empty{}, nil
}
