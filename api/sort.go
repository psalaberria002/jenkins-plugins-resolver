package api

// ByPluginName implements the sort.Interface interface for sorting list of plugins
type ByPluginName []*Plugin

func (pl ByPluginName) Len() int {
	return len(pl)
}

func (pl ByPluginName) Less(i, j int) bool {
	return pl[i].Name < pl[j].Name
}

func (pl ByPluginName) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}

// ByPluginMetadataName implements the sort.Interface interface for sorting list of plugins
type ByPluginMetadataName []*PluginMetadata

func (pl ByPluginMetadataName) Len() int {
	return len(pl)
}

func (pl ByPluginMetadataName) Less(i, j int) bool {
	return pl[i].Plugin.Name < pl[j].Plugin.Name
}

func (pl ByPluginMetadataName) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}
