package service

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	pb "ryg-user-service/gen_proto/user_service"
	"ryg-user-service/model"
)

var testDb *gorm.DB

func setupDatabase() {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "user",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	postgresContainer, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	host, _ := postgresContainer.Host(context.Background())
	port, _ := postgresContainer.MappedPort(context.Background(), "5432")

	dsn := fmt.Sprintf("host=%s port=%s user=user password=password dbname=testdb sslmode=disable", host, port.Port())

	// Retry connection to ensure the database is fully ready
	for i := 0; i < 5; i++ {
		testDb, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}

	if testDb == nil {
		log.Fatalf("failed to connect to database after multiple attempts")
	}

	// Migrate the User model to create tables
	if err := testDb.AutoMigrate(&model.User{}); err != nil {
		log.Fatalf("failed to migrate database: %s", err)
	}
}

func TestMain(m *testing.M) {
	setupDatabase()
	defer testDb.Exec("DROP TABLE users;") // Clean up the database after tests

	m.Run()
}

var user = model.User{
	FullName: "Test User",
	Email:    "testuser@example.com",
	Password: "password123",
	Role:     "user",
	IsActive: true,
}

func TestCreateUser(t *testing.T) {
	defer clearDatabase()
	userService := &UserService{DB: testDb}

	req := &pb.CreateUserRequest{
		FullName: "Test User",
		Email:    "testuser@example.com",
		Password: "password123",
	}

	resp, err := userService.CreateUser(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.FullName, resp.FullName)
	assert.Equal(t, req.Email, resp.Email)
	assert.Equal(t, "user", resp.Role)
}

func clearDatabase() {
	tables, _ := testDb.Migrator().GetTables()
	for _, table := range tables {
		testDb.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", table))
	}
}

func TestGetUserById(t *testing.T) {
	defer clearDatabase()
	userService := &UserService{DB: testDb}

	// First create a user for testing
	testDb.Create(&user)

	req := &pb.GetUserRequest{Id: user.ID}

	resp, err := userService.GetUserById(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, user.ID, resp.Id)
	assert.Equal(t, user.FullName, resp.FullName)
}

func TestUpdateUser(t *testing.T) {
	userService := &UserService{DB: testDb}

	// First create a user for testing
	testDb.Create(&user)

	req := &pb.UpdateUserRequest{
		Id:       user.ID,
		FullName: "Updated User",
	}

	resp, err := userService.UpdateUser(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.FullName, resp.FullName)
}

func TestDeleteUser(t *testing.T) {
	userService := &UserService{DB: testDb}

	// First create a user for testing
	testDb.Create(&user)

	req := &pb.DeleteUserRequest{Id: user.ID}

	resp, err := userService.DeleteUser(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, &emptypb.Empty{}, resp)

	// Verify user is deleted
	var deletedUser model.User
	result := testDb.First(&deletedUser, user.ID)
	assert.Error(t, result.Error)
	assert.True(t, errors.Is(result.Error, gorm.ErrRecordNotFound))
}
