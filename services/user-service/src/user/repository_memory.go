package user

import "sync"

// memoryRepository is a thread-safe in-memory implementation of Repository.
// Used in tests and local development without a database.
type memoryRepository struct {
	mu    sync.RWMutex
	users map[string]*User
}

// NewMemoryRepository returns a new instance of memoryRepository.
func NewMemoryRepository() Repository {
	return &memoryRepository{users: make(map[string]*User)}
}

// Create adds a new user to the in-memory store.
func (r *memoryRepository) Create(u *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	return nil
}

// Get retrieves a user by ID from the in-memory store.
func (r *memoryRepository) Get(id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

// List returns all users from the in-memory store.
func (r *memoryRepository) List() ([]*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make([]*User, 0, len(r.users))
	for _, u := range r.users {
		users = append(users, u)
	}
	return users, nil
}

// Update modifies an existing user in the in-memory store.
func (r *memoryRepository) Update(u *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[u.ID]; !ok {
		return ErrUserNotFound
	}
	r.users[u.ID] = u
	return nil
}

// Delete removes a user from the in-memory store by ID.
func (r *memoryRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[id]; !ok {
		return ErrUserNotFound
	}
	delete(r.users, id)
	return nil
}
