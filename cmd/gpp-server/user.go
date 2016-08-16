package main

import (
	"github.com/fangdingjun/gpp/util"
	"log"
	"os/user"
	"strconv"
)

func dropPrivilege() {
	// go1.7 will add user.LookupGroup, can't use it now

	// if cfg.Group != "" {
	//	 g, err := user.LookupGroup(cfg.Group)
	//	 gid, _:= strconv.Atoi(g.Gid)
	//	 err = util.Setgid(gid)
	// }

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
