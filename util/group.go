// +build linux,cgo darwin,cgo

package util

/*
#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
#include <grp.h>

int get_gid_by_name(char *name){
	struct group *grp;
	grp = getgrnam(name);
	if (grp != NULL){
		return grp->gr_gid;
	}else{
		return -1;
	}
}

char * get_name_by_gid(int gid){
	struct group *grp;
	grp = getgrgid(gid);
	if (grp == NULL){
		return NULL;
	}
	return grp->gr_name;
}

*/
import "C"

import (
	"unsafe"
)

// Group is group struct
type Group struct {
	Gid  int
	Name string
}

// LookupGroupID return a Group by the group id
func LookupGroupID(gid int) *Group {
	name := C.get_name_by_gid(C.int(gid))
	n := C.GoString(name)
	if n == "" {
		return nil
	}
	return &Group{gid, n}
}

// LookupGroup return a Group by the group name
func LookupGroup(name string) *Group {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	gid := C.get_gid_by_name(n)
	if int(gid) == -1 {
		return nil
	}

	return &Group{int(gid), name}
}
