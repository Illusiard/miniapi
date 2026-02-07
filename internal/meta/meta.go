package meta

type Field struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}

type Entity struct {
	Name   string  `json:"name"`
	Table  string  `json:"table"`
	Fields []Field `json:"fields"`
	Module string  `json:"module"`
}

type Registry struct {
	entities []Entity
	modules  []Module
}

func New() *Registry {
	return &Registry{
		entities: make([]Entity, 0, 16),
		modules:  make([]Module, 0, 16),
	}
}

func (r *Registry) AddEntity(e Entity) {
	r.entities = append(r.entities, e)
}

func (r *Registry) Entities() []Entity {
	out := make([]Entity, len(r.entities))
	copy(out, r.entities)
	return out
}


func (r *Registry) AddModule(m Module) {
	r.modules = append(r.modules, m)
}

func (r *Registry) Modules() []Module {
	out := make([]Module, len(r.modules))
	copy(out, r.modules)
	return out
}
