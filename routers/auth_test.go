package routers

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestProcessAuthConfig(t *testing.T) {
	tests := []struct {
		name     string
		authList []string
		accounts gin.Accounts
	}{
		{"basic 1", []string{"user=password"}, gin.Accounts{"user": "password"}},
		{"two", []string{"user=password", "user1=password1"}, gin.Accounts{"user": "password", "user1": "password1"}},
	}

	for idx, test := range tests {
		t.Run(fmt.Sprintf("test #%d: %s", idx+1, test.name), func(t *testing.T) {
			a := processAuthConfig(test.authList)
			assert.Equal(t, test.accounts, a)
		})
	}
}
