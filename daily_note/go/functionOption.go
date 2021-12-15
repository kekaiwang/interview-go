package main

import "fmt"

type Option func(*User)

type User struct {
	Name    string `json:"name"`
	Age     int32  `json:"age"`
	Country string `json:"country"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
}

func loadOptions(opts ...Option) *User {
	user := &User{}

	for _, option := range opts {
		option(user)
	}

	return user
}

func NewUser(u *User, opts ...Option) *User {
	user := u

	for _, option := range opts {
		option(user)
	}

	return user
}

func WithOption(user User) Option {
	return func(u *User) {
		*u = user
	}
}

func WithName(name string) Option {
	return func(u *User) {
		u.Name = name
	}
}

func WithAge(age int32) Option {
	return func(u *User) {
		u.Age = age
	}
}

func main() {
	u := NewUser(&User{
		Name:    "test",
		Country: "BeiJing",
	}, WithAge(18))

	fmt.Println(u)

	opts := WithOption(User{
		Name:    "optino",
		Age:     19,
		Country: "Shanghai",
	})

	u = loadOptions(opts)
	fmt.Println(u)
}
