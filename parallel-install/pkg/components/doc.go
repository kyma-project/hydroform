//Package components defines the installation-related contract for Kyma components.
//A Component is defined by its name, namespace, Helm's chart directory in a local filesystem and a set of overrides.
//A properly defined Component can be installed or uninstalled as a Helm release.
//
//The code in the package uses user-provided function for logging.
package components
