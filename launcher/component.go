package launcher

import (
	"bytes"
	"os/exec"
)

type Component struct {
	Name   string `json:"name"`
	cmd    *exec.Cmd
	output bytes.Buffer
}

func NewComponent(name string) *Component {
	component := Component{Name: name}
	return &component
}
