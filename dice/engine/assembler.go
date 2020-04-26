package engine



type AssemblerCore interface {
	GenerateSuper()([]byte, error)
	PullTile(name string, version string)([]byte, error)
	GenerateCdkApp()([]byte, error)

}