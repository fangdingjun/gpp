package main

import (
	"github.com/fangdingjun/gpp/util"
	"log"
	"os/user"
	"strconv"
)

func dropPrivilege() {
	// go1.7 will add user.LookupGroup
	// now use ourself LookupGroup

	if cfg.Group != "" {
		g := util.LookupGroup(cfg.Group)
		if g != nil {
			err := util.Setgid(g.Gid)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Printf("group %s does not exists\n", cfg.Group)
		}
	}

	if cfg.User != "" {
		u, err := user.Lookup(cfg.User)
		if err != nil {
			log.Printf("user %s does not exists\n", cfg.User)
			return
		}
		uid, _ := strconv.Atoi(u.Uid)
		err = util.Setuid(uid)
		if err != nil {
			log.Println(err)
		}

	}
}
