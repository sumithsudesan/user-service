package user

import (
	"time"

	"github.com/sumithsudesan/pkg/logger"
)

var (
	ErrNameRequired    = errorf("name is required")
	ErrEmailRequired   = errorf("email is required")
	ErrUserNotFound    = errorf("user not found")
	ErrVersionMismatch = errorf("version mismatch")
	ErrInvalidRequest  = errorf("invalid request")
)

type domainError string

func (e domainError) Error() string { return string(e) }
func errorf(s string) error         { return domainError(s) }

// Service implements user domain business logic.
// It depends only on the DB and Publisher interfaces —
type Service struct {
	repo Repository
	pub  Publisher
	log  logger.Logger
}

// Creates new instance of user service
func NewService(repo Repository,
	pub Publisher,
	log logger.Logger) *Service {
	return &Service{repo: repo,
		pub: pub,
		log: log}
}

// Create creates a new user and publishes a "user.created" event.
func (s *Service) Create(input CreateInput) (*User, error) {
	if input.Name == "" {
		return nil, ErrNameRequired
	}
	if input.Email == "" {
		return nil, ErrEmailRequired
	}

	now := time.Now().UTC()
	u := &User{
		ID:        "user-" + now.Format("20060102150405.000000000"),
		Name:      input.Name,
		Email:     input.Email,
		Status:    input.Status,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}

	// create in DB
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}

	// publish user crae event
	_ = s.pub.Publish(Event{
		UserID:    u.ID,
		Email:     u.Email,
		EventType: "user.created",
		Timestamp: now,
	})

	s.log.Info("user created", "user_id", u.ID)
	return u, nil
}

// Used to get user
func (s *Service) Get(id string) (*User, error) {
	// Get user
	return s.repo.Get(id)
}

// Used to,list user
func (s *Service) List() ([]*User, error) {
	return s.repo.List()
}

// used to update user
func (s *Service) Update(id string, input UpdateInput) (*User, error) {
	u, err := s.repo.Get(id)
	if err != nil {
		return nil, err
	}
	if u.Version != input.Version {
		return nil, ErrVersionMismatch
	}

	u.Name = input.Name
	u.Email = input.Email
	u.Status = input.Status
	u.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(u); err != nil {
		return nil, err
	}

	// Evnet user uodated
	_ = s.pub.Publish(Event{
		UserID:    u.ID,
		Email:     u.Email,
		EventType: "user.updated",
		Timestamp: u.UpdatedAt,
	})

	s.log.Info("user updated", "user_id", u.ID)
	return u, nil
}

// Used to delete user
func (s *Service) Delete(id string) error {
	u, err := s.repo.Get(id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	// event user deleted
	_ = s.pub.Publish(Event{
		UserID:    u.ID,
		Email:     u.Email,
		EventType: "user.deleted",
		Timestamp: time.Now().UTC(),
	})

	s.log.Info("user deleted", "user_id", id)
	return nil
}
