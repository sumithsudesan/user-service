package user

// Repository defines the storage contract for the user domain.
// DB Impl
type Repository interface {
	Create(u *User) error
	Get(id string) (*User, error)
	List() ([]*User, error)
	Update(u *User) error
	Delete(id string) error
}
