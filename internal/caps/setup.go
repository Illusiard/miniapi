package caps

import "log/slog"

type Setup struct {
	Routes Routes
	Meta   Meta
	Store  Store
	Log    *slog.Logger
}
