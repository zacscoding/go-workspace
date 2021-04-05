package person

import (
	"errors"
	"sync/atomic"
)

var (
	ErrNotFound = errors.New("not found person")
)

// Person represents a person entity.
type Person struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Age  int32  `json:"age"`
}

var (
	idGenerator int64
)

// Repository represents simple Person CRUD in-memory based repository to tests.
type Repository struct {
	m map[int64]*Person
}

func NewRepository() *Repository {
	idGenerator = 3
	return &Repository{
		m: map[int64]*Person{
			1: {ID: 1, Name: "user1", Age: 12},
			2: {ID: 2, Name: "user2", Age: 15},
			3: {ID: 3, Name: "user3", Age: 21},
		},
	}
}

func (r *Repository) Save(p *Person) *Person {
	id := atomic.AddInt64(&idGenerator, 1)
	r.m[id] = &Person{
		ID:   id,
		Name: p.Name,
		Age:  p.Age,
	}
	p.ID = id
	return p
}

func (r *Repository) FindByID(id int64) (*Person, error) {
	p, ok := r.m[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &Person{
		ID:   p.ID,
		Name: p.Name,
		Age:  p.Age,
	}, nil
}

func (r *Repository) FindAllByName(name string) []*Person {
	var ret []*Person
	for _, p := range r.m {
		if p.Name == name {
			ret = append(ret, &Person{
				ID:   p.ID,
				Name: p.Name,
				Age:  p.Age,
			})
		}
	}
	return ret
}

func (r *Repository) DeleteByID(id int64) (*Person, error) {
	p, ok := r.m[id]
	if !ok {
		return nil, ErrNotFound
	}
	delete(r.m, id)
	return p, nil
}
