package models

import (
	"fmt"

	"github.com/osuAkatsuki/akatsuki-api/common"
)

type Context struct {
	User     SessionUser
	Token    string
	Language string
}

type SessionUser struct {
	ID         int
	Username   string
	Privileges common.UserPrivileges
}

// OnlyUserPublic returns a string containing "(user.privileges & 1 = 1 OR users.id = <userID>)"
// if the user does not have the UserPrivilege AdminManageUsers, and returns "1" otherwise.
func (ctx Context) OnlyUserPublic() string {
	if ctx.User.Privileges&common.AdminPrivilegeManageUsers == common.AdminPrivilegeManageUsers {
		return "1"
	}
	// It's safe to use sprintf directly even if it's a query, because UserID is an int.
	return fmt.Sprintf("(users.privileges & 1 = 1 OR users.id = '%d')", ctx.User.ID)
}
