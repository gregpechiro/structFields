package structFields

import "go/types"

type Generator interface {
	Qf(pkg *types.Package) string
}
