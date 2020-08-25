package main

// curl -XGET http://localhost:3000/persons --header 'trace-id: a'
// curl -XGET http://localhost:3000/persons
func main() {
	s := NewServer()
	if err := s.Run(":3000"); err != nil {
		panic(err)
	}
}
