// +build !cgo

package util

// Group is group struct
type Group struct {
	Gid  int
	Name string
}

// LookupGroupID return a Group by the group id
func LookupGroupID(gid int) *Group {
	return nil
}

// LookupGroup return a Group by the group name
func LookupGroup(name string) *Group {
	return nil
}
