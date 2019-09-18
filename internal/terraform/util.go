package terraform

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
