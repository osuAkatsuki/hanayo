package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"zxq.co/ripple/rippleapi/common"
)

type context struct {
	User     sessionUser
	Token    string
	Language string
}
type sessionUser struct {
	ID         int
	Username   string
	Privileges common.UserPrivileges
	Flags      uint64
}

// OnlyUserPublic returns a string containing "(user.privileges & 1 = 1 OR users.id = <userID>)"
// if the user does not have the UserPrivilege AdminManageUsers, and returns "1" otherwise.
func (ctx context) OnlyUserPublic() string {
	if ctx.User.Privileges&common.AdminPrivilegeManageUsers == common.AdminPrivilegeManageUsers {
		return "1"
	}
	// It's safe to use sprintf directly even if it's a query, because UserID is an int.
	return fmt.Sprintf("(users.privileges & 1 = 1 OR users.id = '%d')", ctx.User.ID)
}

func getContext(c *gin.Context) context {
	return c.MustGet("context").(context)
}
