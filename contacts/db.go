package contacts

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/exp/maps"
)

const pageSize = 100
const dbJsonFileName = "contacts.json"

// Db is a database of contacts
type Db struct {
	contactsByID map[int]*Contact
	dbFile       string
}

// NewDb returns a new Db instance
func NewDb() *Db {
	return &Db{
		contactsByID: make(map[int]*Contact),
		dbFile:       dbJsonFileName,
	}
}

// ValidateContact validates a contact
func (cd *Db) ValidateContact(c *Contact) bool {
	if len(c.Email) == 0 {
		c.Errors["email"] = "Email is required"
	}
	var existingContact *Contact
	for _, contact := range cd.contactsByID {
		if contact.ID != c.ID && contact.Email == c.Email {
			existingContact = contact
		}
	}
	if existingContact != nil {
		c.Errors["email"] = "Email must be unique"
	}
	return len(c.Errors) == 0
}

// SaveContact saves a contact to the contactsByID map and saves it to the JSON file
func (cd *Db) SaveContact(c *Contact) error {
	if !cd.ValidateContact(c) {
		return nil
	}

	if c.ID == 0 {
		c.ID = len(cd.contactsByID) + 1
	}
	cd.contactsByID[c.ID] = c

	err := cd.SaveDB()
	if err != nil {
		return fmt.Errorf("error saving contact DB: %w", err)
	}
	return nil
}

// DeleteContact deletes a contact from the contactsByID map and saves it to the JSON file
func (cd *Db) DeleteContact(c *Contact) error {
	if _, ok := cd.contactsByID[c.ID]; !ok {
		return fmt.Errorf("contact with id %d not found", c.ID)
	}

	delete(cd.contactsByID, c.ID)

	err := cd.SaveDB()
	if err != nil {
		return fmt.Errorf("error saving contact DB: %w", err)
	}

	return nil
}

// LoadDB loads the contacts from the JSON file into the contactsByID map
func (cd *Db) LoadDB() error {
	contacts := make([]*Contact, 0)
	jsonData, err := os.ReadFile(cd.dbFile)
	if err != nil {
		return fmt.Errorf("error reading contactsDb: %w", err)
	}

	err = json.Unmarshal(jsonData, &contacts)
	if err != nil {
		return fmt.Errorf("error unmarshalling contactsDb: %w", err)
	}
	contactsDb := make(map[int]*Contact)
	for _, c := range contacts {
		contactsDb[c.ID] = c
	}
	cd.contactsByID = contactsDb
	return nil
}

// SaveDB saves the contacts from the contactsByID map to the JSON file
func (cd *Db) SaveDB() error {
	contacts := maps.Values(cd.contactsByID)
	jsonData, err := json.MarshalIndent(contacts, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling contacts: %w", err)
	}

	err = os.WriteFile(cd.dbFile, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing contacts: %w", err)
	}

	return nil
}

// Count returns the number of contacts in the contactsByID map
func (cd *Db) Count() int {
	return len(cd.contactsByID)
}

// All returns all contacts in the contactsByID map for the given page
func (cd *Db) All(page int) ([]*Contact, error) {
	start := (page - 1) * pageSize
	end := start + pageSize
	contacts := maps.Values(cd.contactsByID)
	if start >= len(contacts) {
		return nil, fmt.Errorf("page %d not found", page)
	}
	if end > len(contacts) {
		end = len(contacts)
	}
	return contacts[start:end], nil
}

// Search returns all contacts in the contactsByID map that contain the given text
func (cd *Db) Search(text string) ([]*Contact, error) {
	var contacts []*Contact
	for _, c := range cd.contactsByID {
		if strings.Contains(c.First, text) ||
			strings.Contains(c.Last, text) ||
			strings.Contains(c.Email, text) ||
			strings.Contains(c.Phone, text) {
			contacts = append(contacts, c)
		}
	}
	return contacts, nil
}

// Find returns the contact with the given id
func (cd *Db) Find(id int) (*Contact, error) {
	if c, ok := cd.contactsByID[id]; ok {
		c.Errors = make(map[string]string)
		return c, nil
	}
	return nil, fmt.Errorf("contact with id %d not found", id)
}
