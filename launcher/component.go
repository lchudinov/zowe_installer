package launcher

import (
	"os/exec"
)

type Component struct {
	Name   string `json:"name"`
	cmd    *exec.Cmd
	output *Buffer
}

func NewComponent(name string) *Component {
	return &Component{
		Name:   name,
		output: NewBuffer(),
	}
}
