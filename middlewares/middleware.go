package middlewares

import (
	"myapp/data"

	"github.com/lozhkindm/celeritas"
)

type Middleware struct {
	App    *celeritas.Celeritas
	Models data.Models
}
