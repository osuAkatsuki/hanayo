package main

import (
	"strings"
)

// stupid file for gulag privileges
const (
	NORMAL Privileges = 1 << iota // unbanned
	VERIFIED // has logged in

	WHITELISTED // can bypass some anticheat measures

	SUPPORTER // has supporter
	PREMIUM // has premium

	_

	ALUMNI // notable user, mostly ex-staff

	_
	_

	TOURNAMENT // can manage matches without host
	NOMINATOR // can manage map statuses
	MODERATOR // level 1 of being able to manage users
	ADMINISTRATOR // level 2 of being able to manage users
	DEVELOPER // can manage the server's entire state

	DONATOR Privileges = SUPPORTER | PREMIUM
	STAFF Privileges = MODERATOR | ADMINISTRATOR | DEVELOPER
)

type Privileges uint64

var privilegeString = [...]string{
	"Normal",
	"Verified",
	"Whitelisted",
	"Supporter",
	"Premium",
	"Alumni",
	"Tournament",
	"Nominator",
	"Moderator",
	"Administrator",
	"Developer",
	"Donator",
	"Staff",
}

func (p Privileges) String() string {
	var pvs []string
	for i, v := range privilegeString {
		if uint64(p)&uint64(1<<uint(i)) != 0 {
			pvs = append(pvs, v)
		}
	}
	return strings.Join(pvs, ", ")
}