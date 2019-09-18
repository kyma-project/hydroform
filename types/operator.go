package types

import "github.com/kyma-incubator/hydroform/internal/terraform"

type OperatorType string

const (
	Terraform OperatorType = "terraform"
)

type OperatorState = terraform.State
