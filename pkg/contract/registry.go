package contract

type PluginFactory func(PluginFactoryContext) ([]Plugin, error)

type PluginFactoryContext struct {
	ProfileID    string
	Capabilities []Capability
	Config       map[string]string
	Store        Store
	Deps         map[string]any
}
