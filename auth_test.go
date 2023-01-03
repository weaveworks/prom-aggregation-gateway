package main

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestProcessAuthConfig(t *testing.T) {
	tests := []struct {
		name       string
		authString string
		accounts   gin.Accounts
	}{
		{"basic 1", "user=password", gin.Accounts{"user": "password"}},
		{"two", "user=password,user1=password1", gin.Accounts{"user": "password", "user1": "password1"}},
	}

	for idx, test := range tests {
		t.Run(fmt.Sprintf("test #%d: %s", idx+1, test.name), func(t *testing.T) {
			a := processAuthConfig(test.authString)
			assert.Equal(t, test.accounts, a)
		})
	}
}
