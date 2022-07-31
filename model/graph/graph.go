package graph

import (
	"fmt"

	"rxdrag.com/entify/model/domain"
	"rxdrag.com/entify/model/meta"
	"rxdrag.com/entify/model/table"
)

type Model struct {
	Enums        []*Enum
	Interfaces   []*Interface
	Entities     []*Entity
	Services     []*Service
	ValueObjects []*Class
	Relations    []*Relation
	Tables       []*table.Table
}

func New(m *domain.Model) *Model {
	model := Model{}

	for i := range m.Enums {
		model.Enums = append(model.Enums, NewEnum(m.Enums[i]))
	}

	//构建所有接口
	for i := range m.Classes {
		cls := m.Classes[i]
		if cls.StereoType == meta.CLASSS_ABSTRACT {
			model.Interfaces = append(model.Interfaces, NewInterface(cls))
		}
	}

	for i := range m.Classes {
		cls := m.Classes[i]
		if cls.StereoType == meta.CLASSS_ENTITY {
			newEntity := NewEntity(cls)
			model.Entities = append(model.Entities, newEntity)
			//构建接口实现关系
			allParents := cls.AllParents()
			for j := range allParents {
				parentInterface := model.GetInterfaceByUuid(allParents[j].Uuid)
				if parentInterface == nil {
					panic("Can not find interface by uuid:" + allParents[j].Uuid)
				}
				parentInterface.Children = append(parentInterface.Children, newEntity)
				newEntity.Interfaces = append(newEntity.Interfaces, parentInterface)
			}
		} else if cls.StereoType == meta.CLASSS_ABSTRACT {
			allParents := cls.AllParents()
			intf := model.GetInterfaceByUuid(cls.Uuid)
			if intf == nil {
				panic("Can not find interface by uuid:" + cls.Uuid)
			}
			for j := range allParents {
				parentInterface := model.GetInterfaceByUuid(allParents[j].Uuid)
				if parentInterface == nil {
					panic("Can not find interface by uuid:" + allParents[j].Uuid)
				}
				intf.Parents = append(intf.Parents, parentInterface)
			}
		} else if cls.StereoType == meta.CLASS_VALUE_OBJECT {
			model.ValueObjects = append(model.ValueObjects, NewClass(cls))
		} else if cls.StereoType == meta.CLASS_SERVICE {
			model.Services = append(model.Services, NewService(cls))
		}
	}

	//处理关联
	for i := range m.Relations {
		relation := m.Relations[i]
		//增加派生关联
		model.makeRelation(relation)
	}

	//处理属性的实体类型跟枚举类型
	for i := range model.Interfaces {
		intf := model.Interfaces[i]
		model.makeInterface(intf)

	}

	for i := range model.Entities {
		ent := model.Entities[i]
		model.makeEntity(ent)
	}

	//处理Table
	for i := range model.Entities {
		ent := model.Entities[i]
		model.Tables = append(model.Tables, NewEntityTable(ent))
	}

	for i := range model.Relations {
		relation := model.Relations[i]
		model.Tables = append(model.Tables, NewRelationTable(relation))
	}

	return &model
}

func (m *Model) makeRelation(relation *domain.Relation) {
	sourceEntity := m.GetEntityByUuid(relation.Source.Uuid)

	if sourceEntity == nil {
		panic("Can not find souce by uuid:" + relation.Source.Uuid)
	}
	targetEntity := m.GetEntityByUuid(relation.Target.Uuid)

	if targetEntity == nil {
		panic("Can not find target by uuid:" + relation.Target.Uuid)
	}
	r := NewRelation(
		relation,
		sourceEntity,
		targetEntity,
	)
	m.Relations = append(m.Relations, r)

	sourceEntity.AddAssociation(NewAssociation(r, sourceEntity.Uuid()))
	if relation.RelationType == meta.TWO_WAY_AGGREGATION ||
		relation.RelationType == meta.TWO_WAY_ASSOCIATION ||
		relation.RelationType == meta.TWO_WAY_COMBINATION {
		targetEntity.AddAssociation(NewAssociation(r, targetEntity.Uuid()))
	}
}

func (m *Model) makeInterface(intf *Interface) {
	for j := range intf.attributes {
		attr := intf.attributes[j]
		if attr.Type == meta.ENUM || attr.Type == meta.ENUM_ARRAY {
			attr.EumnType = m.GetEnumByUuid(attr.TypeUuid)
		}

		if attr.Type == meta.ENTITY || attr.Type == meta.ENTITY_ARRAY {
			attr.EnityType = m.GetEntityByUuid(attr.TypeUuid)
		}

		if attr.Type == meta.VALUE_OBJECT || attr.Type == meta.VALUE_OBJECT_ARRAY {
			attr.ValueObjectType = m.GetValueObjectByUuid(attr.TypeUuid)
		}
	}
	for j := range intf.methods {
		method := intf.methods[j]
		if method.Method.Type == meta.ENUM || method.Method.Type == meta.ENUM_ARRAY {
			method.EumnType = m.GetEnumByUuid(method.Method.TypeUuid)
		}

		if method.Method.Type == meta.ENTITY || method.Method.Type == meta.ENTITY_ARRAY {
			method.EnityType = m.GetEntityByUuid(method.Method.TypeUuid)
		}

		if method.Method.Type == meta.VALUE_OBJECT || method.Method.Type == meta.VALUE_OBJECT_ARRAY {
			method.ValueObjectType = m.GetValueObjectByUuid(method.Method.TypeUuid)
		}
	}
}

func (m *Model) makeEntity(ent *Entity) {
	for j := range ent.attributes {
		attr := ent.attributes[j]
		if attr.Type == meta.ENUM || attr.Type == meta.ENUM_ARRAY {
			attr.EumnType = m.GetEnumByUuid(attr.TypeUuid)
		}

		if attr.Type == meta.ENTITY || attr.Type == meta.ENTITY_ARRAY {
			attr.EnityType = m.GetEntityByUuid(attr.TypeUuid)
		}

		if attr.Type == meta.VALUE_OBJECT || attr.Type == meta.VALUE_OBJECT_ARRAY {
			attr.ValueObjectType = m.GetValueObjectByUuid(attr.TypeUuid)
		}
	}
	for j := range ent.methods {
		method := ent.methods[j]
		if method.Method.Type == meta.ENUM || method.Method.Type == meta.ENUM_ARRAY {
			method.EumnType = m.GetEnumByUuid(method.Method.TypeUuid)
		}

		if method.Method.Type == meta.ENTITY || method.Method.Type == meta.ENTITY_ARRAY {
			method.EnityType = m.GetEntityByUuid(method.Method.TypeUuid)
		}

		if method.Method.Type == meta.VALUE_OBJECT || method.Method.Type == meta.VALUE_OBJECT_ARRAY {
			method.ValueObjectType = m.GetValueObjectByUuid(method.Method.TypeUuid)
		}
	}
}

func (m *Model) Validate() {
	//检查空实体（除ID外没有属性跟关联）
	for _, entity := range m.Entities {
		if entity.IsEmperty() {
			panic(fmt.Sprintf("Entity %s should have one normal field at least", entity.Name()))
		}
	}
	for _, entity := range m.Services {
		if entity.IsEmperty() {
			panic(fmt.Sprintf("Entity %s should have one normal field at least", entity.Name()))
		}
	}
}

func (m *Model) RootEnities() []*Entity {
	entities := []*Entity{}
	for i := range m.Entities {
		ent := m.Entities[i]
		if ent.Domain.Root {
			entities = append(entities, ent)
		}
	}

	return entities
}

func (m *Model) RootInterfaces() []*Interface {
	interfaces := []*Interface{}
	for i := range m.Interfaces {
		intf := m.Interfaces[i]
		if intf.Domain.Root {
			interfaces = append(interfaces, intf)
		}
	}

	return interfaces
}

func (m *Model) GetInterfaceByUuid(uuid string) *Interface {
	for i := range m.Interfaces {
		intf := m.Interfaces[i]
		if intf.Uuid() == uuid {
			return intf
		}
	}
	return nil
}

func (m *Model) GetEntityByUuid(uuid string) *Entity {
	for i := range m.Entities {
		ent := m.Entities[i]
		if ent.Uuid() == uuid {
			return ent
		}
	}
	return nil
}

func (m *Model) GetServiceByUuid(uuid string) *Service {
	for i := range m.Services {
		partial := m.Services[i]
		if partial.Uuid() == uuid {
			return partial
		}
	}
	return nil
}

func (m *Model) GetValueObjectByUuid(uuid string) *Class {
	for i := range m.ValueObjects {
		ent := m.ValueObjects[i]
		if ent.Uuid() == uuid {
			return ent
		}
	}
	return nil
}

func (m *Model) GetInterfaceByName(name string) *Interface {
	for i := range m.Interfaces {
		intf := m.Interfaces[i]
		if intf.Name() == name {
			return intf
		}
	}
	return nil
}

func (m *Model) GetEntityByName(name string) *Entity {
	for i := range m.Entities {
		ent := m.Entities[i]
		if ent.Name() == name {
			return ent
		}
	}
	return nil
}

func (m *Model) GetServiceByName(name string) *Service {
	for i := range m.Services {
		partial := m.Services[i]
		if partial.Name() == name {
			return partial
		}
	}
	return nil
}

func (m *Model) GetMetaEntity() *Entity {
	return m.GetEntityByUuid(meta.MetaClass.Uuid)
}

func (m *Model) GetEnumByUuid(uuid string) *Enum {
	for i := range m.Enums {
		enum := m.Enums[i]
		if enum.Uuid == uuid {
			return enum
		}
	}
	return nil
}
