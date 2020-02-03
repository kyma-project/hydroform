package terraform

import (
	"github.com/pkg/errors"
)

type HydroUI struct {
	errs []error
}

// Ask asks the user for input using the given query. For Hydroform,
// it always responds "yes" to skip terraform confirmation prompts
func (h *HydroUI) Ask(string) (string, error) {
	return "yes", nil
}

// AskSecret asks the user for input using the given query, but does not echo
// the keystrokes to the terminal.
func (h *HydroUI) AskSecret(string) (string, error) {
	return "", nil
}

// Output is called for normal standard output.
// Terraform output is ignored in Hydroform
func (h *HydroUI) Output(string) {}

// Info is called for information related to the previous output.
// In general this may be the exact same as Output, but this gives
// Ui implementors some flexibility with output formats.
// Terraform info is ignored in Hydroform.
func (h *HydroUI) Info(string) {}

// Error saves error messages from terraform as an error slice to be retrieved later by Hydroform.
func (h *HydroUI) Error(s string) {
	h.errs = append(h.errs, errors.New(s))
}

// Warn saves warning messages from terraform as an error slice to be retrieved later by Hydroform.
func (h *HydroUI) Warn(s string) {
	h.errs = append(h.errs, errors.New(s))
}

// Errors returns any errors or warnings that happened during a terraform command execution
func (h *HydroUI) Errors() []error {
	return h.errs
}
