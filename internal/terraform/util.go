package terraform

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/hashicorp/terraform/terraform"
)

// BindVars binds the map of variables to the Platform variables, to be used
// by Terraform
func (p *Platform) BindVars(vars map[string]interface{}) *Platform {
	for name, value := range vars {
		p.Var(name, value)
	}

	return p
}

// Var set a variable with it's value
func (p *Platform) Var(name string, value interface{}) *Platform {
	if len(p.Vars) == 0 {
		p.Vars = make(map[string]interface{})
	}
	p.Vars[name] = value

	return p
}

// WriteState takes a io.Writer as input to write the Terraform state
func (p *Platform) WriteState(w io.Writer) (*Platform, error) {
	return p, terraform.WriteState(p.State, w)
}

// ReadState takes a io.Reader as input to read from it the Terraform state
func (p *Platform) ReadState(r io.Reader) (*Platform, error) {
	state, err := terraform.ReadState(r)
	if err != nil {
		return p, err
	}
	p.State = state
	return p, nil
}

// WriteStateToFile save the state of the Terraform state to a file
func (p *Platform) WriteStateToFile(filename string) (*Platform, error) {
	var state bytes.Buffer
	if _, err := p.WriteState(&state); err != nil {
		return p, err
	}
	return p, ioutil.WriteFile(filename, state.Bytes(), 0644)
}

// ReadStateFromFile will load the Terraform state from a file and assign it to the
// Platform state.
func (p *Platform) ReadStateFromFile(filename string) (*Platform, error) {
	file, err := os.Open(filename)
	if err != nil {
		return p, err
	}
	return p.ReadState(file)
}
