package users

type NoneUsers struct {
}

// Iterate returns the next user in a circular manner
func (users *NoneUsers) Iterate(iteration uint64) *User {
	return &User{}
}

// Validate validates settings
func (users *NoneUsers) Validate() error {
	return nil
}
