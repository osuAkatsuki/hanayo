package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type commit struct {
	Hash          string
	UnixTimestamp time.Time
	Username      string
	Subject       string
	Repository    string
}

func createFromString(s string) (c commit) {
	var r = strings.SplitN(s, "|", 5)
	for i, v := range r {
		switch i {
		case 0:
			c.Hash = v
		case 1:
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				continue
			}
			c.UnixTimestamp = time.Unix(i, 0)
		case 2:
			c.Username = v
		case 3:
			c.Subject = v
		case 4:
			c.Repository = v
		}
	}
	return
}

func loadChangelog(page int) []commit {
	f, err := os.Open(config.MainRippleFolder + "/ci-system/ci-system/changelog.txt")
	defer f.Close()
	if err != nil {
		return nil
	}
	r := bufio.NewScanner(f)

	if page < 0 {
		page = 0
	} else if page > 50000 {
		page = 50000
	}

	times := (page - 1) * 50
	if times < 0 {
		times = 0
	}
	for i := 0; i < times; i++ {
		// Discard n lines
		r.Scan()
	}

	comms := make([]commit, 0, 50)
	for i := 0; r.Scan() && i < 50; i++ {
		s := r.Text()
		comms = append(comms, createFromString(s))
	}
	return comms
}
