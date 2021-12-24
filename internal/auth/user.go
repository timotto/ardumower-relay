package auth

type user struct {
	id string
}

func NewUser(id string) *user {
	return &user{id: id}
}

func (u *user) Id() string {
	return u.id
}
