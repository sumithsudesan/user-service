package user

// CreateInput struct - the input required to create a new user.
type CreateInput struct {
	Name   string
	Email  string
	Status string
}

// UpdateInput struct  - the input required to update an existing user.
type UpdateInput struct {
	Name    string
	Email   string
	Status  string
	Version int
}
