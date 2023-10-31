package pkg

var fixFactories = map[string]func(*Config) block{}

func registerFix() {
	fixFactories["local_file"] = func(c *Config) block {
		return &LocalFile{
			BaseFix: &BaseFix{
				c: c,
			},
		}
	}
}

var ruleFactories = map[string]func(*Config) block{}

func registerRule() {
	ruleFactories["file_hash"] = func(c *Config) block {
		return &FileHashRule{
			BaseRule: &BaseRule{
				c: c,
			},
		}
	}
	ruleFactories["must_be_true"] = func(c *Config) block {
		return &MustBeTrueRule{
			BaseRule: &BaseRule{
				c: c,
			},
		}
	}
}

var datasourceFactories = map[string]func(*Config) block{}

func registerData() {
	datasourceFactories["http"] = func(c *Config) block {
		return &HttpDatasource{
			BaseData: &BaseData{
				c: c,
			},
		}
	}
}
