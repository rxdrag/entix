package script

import (
	"context"
	"log"

	"github.com/graphql-go/graphql"
	"rxdrag.com/entify/logs"
	"rxdrag.com/entify/model/data"
	"rxdrag.com/entify/model/graph"
	"rxdrag.com/entify/modules/register"
	"rxdrag.com/entify/orm"
	"rxdrag.com/entify/service"
)

type ScriptService struct {
	ctx     context.Context
	roleIds []uint64
	model   *graph.Model
	session *orm.Session
}

func NewService(ctx context.Context, model *graph.Model) *ScriptService {

	return &ScriptService{
		ctx:     ctx,
		model:   model,
		roleIds: service.QueryRoleIds(ctx, model),
	}
}

func (s *ScriptService) BeginTx() {
	session, err := orm.Open()
	if err != nil {
		log.Panic(err.Error())
	}
	s.session = session
	err = session.BeginTx()
	if err != nil {
		log.Panic(err.Error())
	}
}

func (s *ScriptService) Commit() {
	if s.session == nil {
		log.Panic("No session to commit")
	}
	err := s.session.Commit()

	if err != nil {
		log.Panic(err.Error())
	}
}

func (s *ScriptService) ClearTx() {
	if s.session == nil {
		log.Panic("No session to ClearTx")
	}
	s.session.ClearTx()
	s.session = nil
}

func (s *ScriptService) Rollback() {
	if s.session == nil {
		log.Panic("No session to Rollback")
	}

	err := s.session.Dbx.Rollback()
	if err != nil {
		log.Panic(err.Error())
	}
	s.session = nil
}

func (s *ScriptService) Save(objects []interface{}, entityName string) []orm.InsanceData {
	entity := s.model.GetEntityByName(entityName)

	if entity == nil {
		log.Panic("Can not find entity by name:" + entityName)
	}

	savedIds := []interface{}{}
	for i := range objects {
		object := objects[i]
		data.ConvertObjectId(object.(map[string]interface{}))
		instance := data.NewInstance(object.(map[string]interface{}), entity)
		obj, err := s.session.SaveOne(instance)
		if err != nil {
			log.Panic(err.Error())
		}
		savedIds = append(savedIds, obj)
	}
	if len(savedIds) > 0 {
		logs.WriteModelLog(s.model, &entity.Class, s.ctx, logs.SET, logs.SUCCESS, "", "script")
		return s.session.QueryByIds(entity, savedIds)
	}

	return []orm.InsanceData{}
}

func (s *ScriptService) SaveOne(object interface{}, entityName string) interface{} {
	entity := s.model.GetEntityByName(entityName)

	if entity == nil {
		log.Panic("Can not find entity by name:" + entityName)
	}

	if object == nil {
		log.Panic("Object to save is nil")
	}

	instance := data.NewInstance(object.(map[string]interface{}), entity)

	id, err := s.session.SaveOne(instance)
	if err != nil {
		log.Panic(err.Error())
	}

	result := s.session.QueryOneById(instance.Entity, id)

	return result
}

func (s *ScriptService) WriteLog(
	operate string,
	result string,
	message string,
) {
	logs.WriteBusinessLog(s.model, s.ctx, operate, result, message)
}

func (s *ScriptService) Query(gql string, variables interface{}) interface{} {
	var newVariables map[string]interface{}

	if variables != nil {
		newVariables = variables.(map[string]interface{})
	}
	params := graphql.Params{
		Schema:         register.GetSchema(s.ctx),
		RequestString:  gql,
		VariableValues: newVariables,
		Context:        context.WithValue(s.ctx, "gql", gql),
	}

	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		log.Printf("failed to execute graphql operation, errors: %+v", r.Errors)
		log.Panic(r.Errors[0].Error())
	}

	return r.Data
}