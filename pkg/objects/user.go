package objects

type User struct {
	ID		int
	Name		string `json:"name"`
	Email		string `json:"email"`
	Space		string `json:"space"`
}
