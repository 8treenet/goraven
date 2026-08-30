// Package plugin provides the plugin system for the GoRaven agent framework.
// Third-party extensions implement the hook interfaces defined here and register
// via the Register function to extend agent behavior.
package plugin

// Plugin is the base interface that all plugins must implement.
type Plugin interface {
	// Name returns the unique identifier for this plugin, e.g. "mycorp/approval".
	Name() string
	// Version returns the semantic version of this plugin, e.g. "1.0.0".
	Version() string
}
