package main

import (
	"github.com/fangdingjun/gpp/util"
	"log"
	"os/user"
	"strconv"
	"syscall"
)

func dropPrivilege() {

	uid := syscall.Getuid()

	if uid != 0 {
		// only root(uid=0) can call setuid
		// not root, skip
		return
	}

	// go1.7 will add user.LookupGroup
	// now use ourself LookupGroup
	if cfg.Group != "" {
		g, err := util.LookupGroup(cfg.Group)
		if err == nil {
			err := util.Setgid(g.Gid)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println(err)
		}
	}

	if cfg.User != "" {
		u, err := user.Lookup(cfg.User)
		if err != nil {
			log.Println(err)
			return
		}
		uid, _ := strconv.Atoi(u.Uid)
		err = util.Setuid(uid)
		if err != nil {
			log.Println(err)
		}

	}
}
