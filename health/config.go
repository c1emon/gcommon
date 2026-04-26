package health

// Config configures the hellofresh health registry (component metadata and
// optional runtime system block in JSON responses).
type Config struct {
	// ServiceName is the logical service name (maps to hellofresh Component.Name).
	ServiceName string
	// Version is an optional release or build version (Component.Version).
	Version string
}
