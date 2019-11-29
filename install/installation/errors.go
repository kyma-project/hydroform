package installation

import "fmt"

type InvalidInstallationStateError struct {
	InstallationState  string
	InstallationStatus string
}

func (e InvalidInstallationStateError) Error() string {
	return fmt.Sprintf("invalid installation state: %s, status: %s", e.InstallationState, e.InstallationStatus)
}

type InstallationError struct {
	ShortMessage string
	ErrorEntries []ErrorEntry
}

type ErrorEntry struct {
	Component   string
	Log         string
	Occurrences int32
}

func (e InstallationError) Error() string {
	return e.ShortMessage
}

func (e InstallationError) Details() string {
	errorLogString := "Installation errors: "

	for _, installError := range e.ErrorEntries {
		errorLogString = fmt.Sprintf("%s\nComponent: %s, Log: %s", errorLogString, installError.Component, installError.Log)
	}

	return errorLogString
}
