package cloud_events1

type Event interface {
	Type() string
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (u *User) Type() string {
	return "user"
}

type Product struct {
	ID    int64 `json:"id"`
	Price int   `json:"price"`
	Stock int   `json:"stock"`
}

func (u *Product) Type() string {
	return "product"
}
