package engine

type BrewerCore interface {
	RunAll()([]byte, error)
	RunCdk()([]byte, error)
	RunKubectl()([]byte, error)
}