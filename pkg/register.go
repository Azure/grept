package pkg

var fixFactories = map[string]func(*Config) block{}

func registerFix() {
	fixFactories["local_file"] = func(c *Config) block {
		return &LocalFileFix{
			baseBlock: &baseBlock{
				c: c,
			},
		}
	}
	fixFactories["rename_file"] = func(c *Config) block {
		return &RenameFile{
			baseBlock: &baseBlock{
				c: c,
			},
		}
	}
	fixFactories["rm_local_file"] = func(c *Config) block {
		return &RmLocalFile{
			baseBlock: &baseBlock{
				c: c,
			},
		}
	}
}

var ruleFactories = map[string]func(*Config) block{}

func registerRule() {
	ruleFactories["file_hash"] = func(c *Config) block {
		return &FileHashRule{
			baseBlock: &baseBlock{
				c: c,
			},
		}
	}
	ruleFactories["must_be_true"] = func(c *Config) block {
		return &MustBeTrueRule{
			baseBlock: &baseBlock{
				c: c,
			},
		}
	}
	ruleFactories["dir_exist"] = func(c *Config) block {
		return &DirExistRule{
			baseBlock: &baseBlock{
				c: c,
			},
		}
	}
}

var datasourceFactories = map[string]func(*Config) block{}

func registerData() {
	datasourceFactories["http"] = func(c *Config) block {
		return &HttpDatasource{
			baseBlock: &baseBlock{
				c: c,
			},
		}
	}
	datasourceFactories["git_ignore"] = func(c *Config) block {
		return &GitIgnoreDatasource{
			baseBlock: &baseBlock{
				c: c,
			},
		}
	}
}
