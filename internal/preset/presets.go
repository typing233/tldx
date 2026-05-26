package preset

var Registry = map[string][]string{
	"popular": {"com", "net", "org", "io"},
	"tech":    {"dev", "io", "ai", "app", "tech"},
	"startup": {"co", "io", "ai", "app", "dev"},
	"short":   {"io", "ai", "co", "me", "to"},
	"web":     {"com", "net", "org", "info", "biz"},
	"new":     {"xyz", "online", "site", "store", "fun"},
	"country": {"us", "uk", "ca", "de", "fr", "jp"},
}

func Get(name string) ([]string, bool) {
	tlds, ok := Registry[name]
	return tlds, ok
}

func List() map[string][]string {
	return Registry
}

func Names() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	return names
}
