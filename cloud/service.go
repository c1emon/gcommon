package cloud

import (
	"fmt"
	"sort"
)

// RemoteService is an instance of a service in a discovery system.
type RemoteService struct {
	// ID is the unique instance ID as registered.
	ID string `json:"id"`
	// Name is the service name as registered.
	Name string `json:"name"`
	// Version is the version of the compiled.
	Version string `json:"version"`
	// Metadata is the kv pair metadata associated with the service instance.
	Metadata map[string]string `json:"metadata"`
	// Tags is string list
	Tags []string
	// Endpoints are endpoint addresses of the service instance.
	// schema:
	//   http://127.0.0.1:8000?isSecure=false
	//   grpc://127.0.0.1:9000?isSecure=false
	Endpoints []*Endpoint `json:"endpoints"`
}

func (i *RemoteService) String() string {
	return fmt.Sprintf("%s-%s", i.Name, i.ID)
}

// Equal returns whether i and o are equivalent.
func (i *RemoteService) Equal(o any) bool {
	if i == nil && o == nil {
		return true
	}

	if i == nil || o == nil {
		return false
	}

	t, ok := o.(*RemoteService)
	if !ok {
		return false
	}

	if len(i.Endpoints) != len(t.Endpoints) {
		return false
	}

	// sort.Strings(i.Endpoints)
	// sort.Strings(t.Endpoints)
	// for j := 0; j < len(i.Endpoints); j++ {
	// 	if i.Endpoints[j] != t.Endpoints[j] {
	// 		return false
	// 	}
	// }

	sort.Strings(i.Tags)
	sort.Strings(t.Tags)
	for j := 0; j < len(i.Tags); j++ {
		if i.Tags[j] != t.Tags[j] {
			return false
		}
	}

	if len(i.Metadata) != len(t.Metadata) {
		return false
	}

	for k, v := range i.Metadata {
		if v != t.Metadata[k] {
			return false
		}
	}

	return i.ID == t.ID && i.Name == t.Name && i.Version == t.Version
}
