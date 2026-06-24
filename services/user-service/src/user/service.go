package user

import (
	"errors"
	"time"
)

// Errors that can be returned by the Service methods.
var (
	ErrNameRequired    = errors.New("name is required")  // errors for missing name
	ErrEmailRequired   = errors.New("email is required") // errors for missing email
	ErrUserNotFound    = errors.New("user not found")    // errors for user not found
	ErrVersionMismatch = errors.New("version mismatch")  // errors for version mismatch
	ErrInvalidRequest  = errors.New("invalid request")   // errors for invalid request
)

// User struct represents a user in the system.
type Service struct {
	users map[string]*User // users is a map that stores users by their ID.
}

// new service creates a new instance of the Service.
func NewService() *Service {
	// Initialize the users map to store users by their ID.
	return &Service{
		users: make(map[string]*User),
	}
}

// Create creates a new user with the provided input
// and returns the created user.
func (s *Service) Create(input CreateInput) (*User, error) {
	// Check if the name is provided
	if input.Name == "" {
		return nil, ErrNameRequired
	}
	// Check if the email is provided
	if input.Email == "" {
		return nil, ErrEmailRequired
	}

	now := time.Now().UTC()

	// Create a new user with a unique ID and the provided input.
	u := &User{
		ID:        "user-" + time.Now().Format("20060102150405.000000000"),
		Name:      input.Name,
		Email:     input.Email,
		Status:    input.Status,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}

	s.users[u.ID] = u
	return u, nil
}

// Get retrieves a user by ID. If the user is not found,
// it returns an error.
func (s *Service) Get(id string) (*User, error) {
	u, ok := s.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

// List returns all users in the service.
func (s *Service) List() ([]*User, error) {
	users := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	return users, nil
}

// Update updates an existing user with the provided input and
// returns the updated user.
func (s *Service) Update(id string, input UpdateInput) (*User, error) {
	u, ok := s.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	if u.Version != input.Version {
		return nil, ErrVersionMismatch
	}
	// Update the user fields with the provided input and
	//  increment the version.
	u.Name = input.Name
	u.Email = input.Email
	u.Status = input.Status
	u.UpdatedAt = time.Now().UTC()
	u.Version++

	s.users[id] = u
	return u, nil
}

// Delete removes a user by ID from the service.
func (s *Service) Delete(id string) error {
	if _, ok := s.users[id]; !ok {
		return ErrUserNotFound
	}
	delete(s.users, id)
	return nil
}
