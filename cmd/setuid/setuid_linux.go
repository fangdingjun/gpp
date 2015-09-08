package setuid
// #include <sys/types.h>
// #include <unistd.h>

import "C"

import (
	"log"
	"os/user"
	"strconv"
	//"syscall"
)

func Setuid(run_user string) {
	if run_user != "" {
		u, err := user.Lookup(run_user)
		if err != nil {
			log.Fatal(err)
		}

		uid, _ := strconv.Atoi(u.Uid)
		gid, _ := strconv.Atoi(u.Gid)

		_, err = C.setgid(C.__gid_t(gid))
		if err != nil {
			log.Fatal(err)
		}

		_, err = C.setuid(C.__uid_t(uid))
		if err != nil {
			log.Fatal(err)
		}
	}
}
