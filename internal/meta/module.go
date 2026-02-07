package meta

type Module struct {
	Name        string `json:"name"`
	WithStore   bool   `json:"withStore"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
}
