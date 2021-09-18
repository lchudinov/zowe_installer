package launcher

import (
	"bytes"
	"os/exec"
)

type Component struct {
	Name   string
	cmd    *exec.Cmd
	output bytes.Buffer
}

func NewComponent(name string) *Component {
	component := Component{Name: name}
	return &component
}
