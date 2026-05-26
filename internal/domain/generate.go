package domain

type GenerateConfig struct {
	Keywords  []string
	Prefixes  []string
	Suffixes  []string
	TLDs      []string
	MaxLength int
}

func Generate(cfg GenerateConfig) []string {
	prefixes := cfg.Prefixes
	if len(prefixes) == 0 {
		prefixes = []string{""}
	}
	suffixes := cfg.Suffixes
	if len(suffixes) == 0 {
		suffixes = []string{""}
	}
	tlds := cfg.TLDs
	if len(tlds) == 0 {
		tlds = []string{"com"}
	}

	var results []string
	for _, prefix := range prefixes {
		for _, keyword := range cfg.Keywords {
			for _, suffix := range suffixes {
				for _, tld := range tlds {
					name := prefix + keyword + suffix
					if name == "" {
						continue
					}
					domain := name + "." + tld
					if cfg.MaxLength > 0 && len(domain) > cfg.MaxLength {
						continue
					}
					results = append(results, domain)
				}
			}
		}
	}
	return results
}
