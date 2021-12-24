package model

//counterfeiter:generate -o fake . User
type User interface {
	Id() string
}
