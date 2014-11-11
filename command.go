package cli

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
)

type byName []*Command

func (b byName) Len() int           { return len(b) }
func (b byName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byName) Less(i, j int) bool { return b[i].Name < b[j].Name }

const (
	IntOptT byte = iota
	StringOptT
	BoolOptT
)

type CmdFunc func(*Command) error

type Command struct {
	Name    string
	Desc    string
	Opts    map[string]interface{}
	Args    []string
	Parent  *Command
	subs    map[string]*Command
	flags   *flag.FlagSet
	runFunc func(*Command) error
	nOpts   int
}

type Option struct {
	Name string
	Desc string
	Val  interface{}
	Type byte
}

func (o Option) register(fs *flag.FlagSet) interface{} {
	switch o.Type {
	case IntOptT:
		def, ok := o.Val.(int)
		if ok {
			i := fs.Int(o.Name, def, o.Desc)
			return interface{}(i)
		}
	case StringOptT:
		def, ok := o.Val.(string)
		if ok {
			i := fs.String(o.Name, def, o.Desc)
			return interface{}(i)
		}
	case BoolOptT:
		def, ok := o.Val.(bool)
		if ok {
			i := fs.Bool(o.Name, def, o.Desc)
			return interface{}(i)
		}
	}
	return nil
}

func IntOpt(name string, def int, desc string) Option {
	return Option{name, desc, def, IntOptT}
}

func StringOpt(name string, def string, desc string) Option {
	return Option{name, desc, def, StringOptT}
}

func BoolOpt(name string, def bool, desc string) Option {
	return Option{name, desc, def, BoolOptT}
}

func New(name string, desc string, cmdFunc CmdFunc) *Command {
	cmd := &Command{Name: name, Desc: desc, runFunc: cmdFunc}

	cmd.flags = flag.NewFlagSet(name, flag.ContinueOnError)
	cmd.flags.SetOutput(ioutil.Discard)

	cmd.Opts = make(map[string]interface{})
	cmd.subs = make(map[string]*Command)

	return cmd
}

func (c *Command) AddOpts(opts ...Option) *Command {
	for _, opt := range opts {
		c.Opts[opt.Name] = opt.register(c.flags)
		c.nOpts += 1
	}
	return c
}

func (c *Command) Run(cmdline []string) error {
	err := c.flags.Parse(cmdline[1:])
	if err != nil {
		c.PrintHelp(err)
		return err
	}

	c.Args = c.flags.Args()

	if len(c.Args) > 0 {
		if sub, ok := c.subs[c.Args[0]]; ok {
			return sub.Run(c.Args)
		}
	}

	return c.runFunc(c)

}

func (c *Command) Subs(subs ...*Command) *Command {
	for _, v := range subs {
		v.Parent = c
		c.subs[v.Name] = v
	}
	return c
}

func (c *Command) StringOpt(name string) (string, bool) {
	opt, ok := c.Opts[name]
	if !ok {
		return "", false
	}

	r, ok := opt.(*string)
	if !ok {
		return "", false
	}
	return *r, true
}

func (c *Command) BoolOpt(name string) (bool, bool) {
	opt, ok := c.Opts[name]
	if !ok {
		return false, false
	}

	r, ok := opt.(*bool)
	if !ok {
		return false, false
	}
	return *r, true
}

func (c *Command) IntOpt(name string) (int, bool) {
	opt, ok := c.Opts[name]
	if !ok {
		return 0, false
	}

	r, ok := opt.(*int)
	if !ok {
		return 0, false
	}
	return *r, true
}

func (c *Command) PrintHelp(err error) {
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println()
	}

	fmt.Printf("Usage: %s [OPTIONS] [COMMAND] [arg...]\n\n", c.Full())

	fmt.Println(c.Desc)
	fmt.Println()

	if c.nOpts > 0 {
		fmt.Println("Options:")
		c.flags.SetOutput(os.Stdout)
		c.flags.PrintDefaults()
		c.flags.SetOutput(ioutil.Discard)

		fmt.Println()
	}

	if len(c.subs) > 0 {
		c.printCommands(0, false)
	}
}

func (c *Command) Full() string {
	if c.Parent == nil {
		return c.Name
	}
	return fmt.Sprintf("%s %s", c.Parent.Full(), c.Name)
}

func (c *Command) RecursiveHelp() {
	fmt.Println(c.Desc)
	fmt.Println()

	fmt.Println("Commands:")
	if len(c.subs) > 0 {
		c.printCommands(0, true)
	}
}

func (c *Command) printCommands(level int, recurse bool) {
	if !recurse {
		fmt.Println("Commands:")
	}
	cmds := make([]*Command, len(c.subs))
	i := 0
	for _, v := range c.subs {
		cmds[i] = v
		i++
	}
	sort.Sort(byName(cmds))
	max := 0
	for _, v := range cmds {
		l := len(v.Name)
		if l > max {
			max = l
		}
	}

	for _, v := range cmds {
		l := len(v.Name)
		for i := 0; i < level; i++ {
			fmt.Print("  ")
		}
		fmt.Printf("  %s", v.Name)
		for i := 0; i < (max-l)+3; i++ {
			fmt.Print(" ")
		}
		fmt.Printf("%s\n", v.Desc)
		if recurse && len(c.subs) > 0 {
			v.printCommands(level+1, true)
		}
	}
}

func HelpOnly(cmd *Command) error {
	cmd.PrintHelp(nil)
	return nil
}

func RecursiveHelp(cmd *Command) error {
	cmd.RecursiveHelp()
	return nil
}
