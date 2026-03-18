package resources

const (
	StateAvailable = iota
	StateStalling
	StateUnAvailable
)

type Resource struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State int    `json:"state"`
}

func NewResource(id string, name string) Resource {
	return Resource{
		ID:    id,
		Name:  name,
		State: StateAvailable,
	}
}

var AvailableResources = []Resource{NewResource("abcd", "Seat 1"), NewResource("efgh", "Seat 2"), NewResource("ijkl", "Seat 3")}
