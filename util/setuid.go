// +build linux,cgo darwin,cgo

package util

/*
#include <sys/types.h>
#include <unistd.h>
*/
import "C"

import (
	"errors"
	"fmt"
)

//Setuid set the uid to uid
func Setuid(uid int) error {
	ret := C.setuid(C.__uid_t(uid))
	if ret == C.int(0) {
		return nil
	}

	msg := fmt.Sprintf("setuid return with status %d", ret)
	return errors.New(msg)
}

//Setgid set the gid to gid
func Setgid(gid int) error {
	ret := C.setgid(C.__gid_t(gid))
	if ret == C.int(0) {
		return nil
	}

	msg := fmt.Sprintf("setgid return with status %d", ret)
	return errors.New(msg)
}
