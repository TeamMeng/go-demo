package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestEntry(t *testing.T) {
	password := "hello#world"
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	err = bcrypt.CompareHashAndPassword(encrypted, []byte(password))
	assert.NoError(t, err)
}
