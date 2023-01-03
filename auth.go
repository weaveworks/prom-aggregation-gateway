package main

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func processAuthConfig(auth string) gin.Accounts {
	authAccounts := gin.Accounts{}
	authList := strings.Split(auth, ",")
	if len(authList) == 0 {
		return authAccounts
	}

	for _, item := range authList {
		i := strings.Split(item, "=")
		if len(i) == 2 {
			authAccounts[i[0]] = i[1]
		}
	}

	return authAccounts
}
