package auth

import (
	"sync-backend/api/auth/dto"
	"sync-backend/api/user"
	"sync-backend/arch/config"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
)

type AuthService interface {
	SignUp(signUpRequest *dto.SignUpRequest) (*dto.SignUpResponse, error)
	GenerateToken(userId string) (string, error)
}

type authService struct {
	network.BaseService
	userService user.UserService
}

func NewAuthService(
	db mongo.Database,
	config *config.Config,
) AuthService {
	return &authService{
		BaseService: network.NewBaseService(),
		userService: user.NewUserService(db),
	}
}

func (s *authService) SignUp(signUpRequest *dto.SignUpRequest) (*dto.SignUpResponse, error) {
	user, err := s.userService.CreateUser(signUpRequest.Email, signUpRequest.Password, signUpRequest.Name, signUpRequest.ProfilePicUrl)
	if err != nil {
		return nil, err
	}
	token, err := s.GenerateToken(user.ID.Hex())
	if err != nil {
		return nil, err
	}
	signUpResponse := dto.NewSignUpResponse(user.ID.Hex(), token)
	return signUpResponse, nil

}

func (s *authService) GenerateToken(userId string) (string, error) {
	// Implement token generation logic here
	// This is just a placeholder implementation
	return "generated_" + userId, nil
}
