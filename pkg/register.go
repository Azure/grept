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
	fixFactories["rename_file"] = func(c *Config) block {
		return &RenameFile{
			BaseFix: &BaseFix{
				c: c,
			},
		}
	}
	fixFactories["rm_local_file"] = func(c *Config) block {
		return &RmLocalFile{
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
	ruleFactories["dir_exist"] = func(c *Config) block {
		return &DirExistRule{
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
	datasourceFactories["git_ignore"] = func(c *Config) block {
		return &GitIgnoreDatasource{
			BaseData: &BaseData{
				c: c,
			},
		}
	}
}
