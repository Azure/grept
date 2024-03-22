package golden

func init() {
	RegisterBaseBlock(func() BlockType {
		return new(BaseData)
	})
	RegisterBaseBlock(func() BlockType { return new(BaseResource) })
	RegisterBlock(new(DummyData))
	RegisterBlock(new(DummyResource))
}
