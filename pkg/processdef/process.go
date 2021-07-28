package processdef

import (
	"fmt"

	"github.com/google/shlex"
)

type Process struct {
	ID   string
	Tags []string
	Name string
	Args []string
	Tpl  string
}

func CreateProcessFromTemplate(id string, tags []string, tpl string) (Process, error) {
	args, err := shlex.Split(tpl)
	if err != nil {
		return Process{}, fmt.Errorf("invalid command: %w", err)
	}
	if len(args) == 0 {
		return Process{}, fmt.Errorf("invalid command")
	}
	return Process{
		ID:   id,
		Tags: tags,
		Name: args[0],
		Args: args[1:],
		Tpl:  tpl,
	}, nil
}

func FilterList(i []Process, f func(Process) bool) []Process {
	r := make([]Process, 0)
	for _, p := range i {
		if f(p) {
			r = append(r, p)
		}
	}
	return r
}
