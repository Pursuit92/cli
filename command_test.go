package cli

import (
	"testing"
)

func TestBuild(t *testing.T) {
	base := New("test", "description", nil).AddOpts(IntOpt("int", 5, "an int"))
	sub1 := New("test2", "desc2", nil)
	base.Subs(sub1)

	sub1.PrintHelp(nil)

}
