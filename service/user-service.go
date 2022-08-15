package service

import (
	"log"

	dto "book-keeper/dtos"
	model "book-keeper/models"
	"book-keeper/repository"

	"github.com/mashingan/smapping"
)

//UserService is a contract.....
type UserService interface {
	Update(user dto.UserUpdateDTO) model.User
	Profile(userID string) model.User
}

type userService struct {
	userRepository repository.UserRepository
}

//NewUserService creates a new instance of UserService
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepository: userRepo,
	}
}

func (service *userService) Update(user dto.UserUpdateDTO) model.User {
	userToUpdate := model.User{}
	err := smapping.FillStruct(&userToUpdate, smapping.MapFields(&user))
	if err != nil {
		log.Fatalf("Failed map %v:", err)
	}
	updatedUser := service.userRepository.UpdateUser(userToUpdate)
	return updatedUser
}

func (service *userService) Profile(userID string) model.User {
	return service.userRepository.ProfileUser(userID)
}
