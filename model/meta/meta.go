package meta

type Model struct {
	Classes   []*ClassMeta
	Relations []*RelationMeta
	Packages  []*PackageMeta
}

func New(m *MetaContent) *Model {
	model := Model{
		Classes:   make([]*ClassMeta, len(m.Classes)),
		Relations: make([]*RelationMeta, len(m.Relations)),
		Packages:  make([]*PackageMeta, len(m.Relations)),
	}

	for i := range m.Packages {
		model.Packages[i] = &m.Packages[i]
	}

	for i := range m.Classes {
		model.Classes[i] = &m.Classes[i]
	}

	for i := range m.Relations {
		model.Relations[i] = &m.Relations[i]
	}
	return &model
}

func (m *Model) GetPackageByUuid(uuid string) *PackageMeta {
	for i := range m.Packages {
		if m.Packages[i].Uuid == uuid {
			return m.Packages[i]
		}
	}
	return nil
}
