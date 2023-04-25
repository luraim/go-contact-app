package contacts

type Contact struct {
	ID     int
	First  string
	Last   string
	Phone  string
	Email  string
	Errors map[string]string
}

func NewContact(first, last, phone, email string) *Contact {
	return &Contact{
		First:  first,
		Last:   last,
		Phone:  phone,
		Email:  email,
		Errors: make(map[string]string),
	}
}
