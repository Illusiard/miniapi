package caps

import "github.com/Illusiard/miniapi/internal/meta"

type Meta interface {
	AddEntity(e meta.Entity)
	AddModule(m meta.Module)
}
