package applicationglobals

import (
	"github.com/kevintavog/findaphoto/findaphotoserver/fieldlogger"
)

type ApplicationGlobals struct {
	FieldLogger *fieldlogger.FieldLogger
}

func newGlobals() *ApplicationGlobals {
	return &ApplicationGlobals{FieldLogger: fieldlogger.New()}
}

// Reset gets called just before a new HTTP request starts calling
// middleware + handlers
func (g *ApplicationGlobals) Reset() {
	g.FieldLogger = fieldlogger.New()
}

// Done gets called after the HTTP request has completed right before
// Context gets put back into the pool
func (g *ApplicationGlobals) Done() {
}
