package resolve

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"rxdrag.com/entify/consts"
	"rxdrag.com/entify/log"
	"rxdrag.com/entify/model"
	"rxdrag.com/entify/model/graph"
	"rxdrag.com/entify/repository"
	"rxdrag.com/entify/utils"
)

func QueryOneInterfaceResolveFn(intf *graph.Interface, model *model.Model) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		defer utils.PrintErrorStack()
		repos := repository.New(model)
		repos.MakeInterfaceAbilityVerifier(p, intf)
		instance := repos.QueryOneInterface(intf, p.Args)
		log.WriteModelLog(model, &intf.Class, p, log.SUCCESS, log.QUERY)
		return instance, nil
	}
}

func QueryInterfaceResolveFn(intf *graph.Interface, model *model.Model) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		defer utils.PrintErrorStack()
		repos := repository.New(model)
		repos.MakeInterfaceAbilityVerifier(p, intf)
		result := repos.QueryInterface(intf, p.Args)
		log.WriteModelLog(model, &intf.Class, p, log.SUCCESS, log.QUERY)
		return result, nil
	}
}

func QueryOneEntityResolveFn(entity *graph.Entity, model *model.Model) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		defer utils.PrintErrorStack()
		repos := repository.New(model)
		repos.MakeEntityAbilityVerifier(p, entity.Uuid())
		instance := repos.QueryOneEntity(entity, p.Args)
		log.WriteModelLog(model, &entity.Class, p, log.SUCCESS, log.QUERY)
		return instance, nil
	}
}

func QueryEntityResolveFn(entity *graph.Entity, model *model.Model) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		defer utils.PrintErrorStack()
		repos := repository.New(model)
		repos.MakeEntityAbilityVerifier(p, entity.Uuid())
		result := repos.QueryEntity(entity, p.Args)
		log.WriteModelLog(model, &entity.Class, p, log.SUCCESS, log.QUERY)
		return result, nil
	}
}

func QueryAssociationFn(asso *graph.Association, model *model.Model) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var (
			source      = p.Source.(map[string]interface{})
			v           = p.Context.Value
			loaders     = v(consts.LOADERS).(*Loaders)
			handleError = func(err error) error {
				return fmt.Errorf(err.Error())
			}
		)
		defer utils.PrintErrorStack()

		if loaders == nil {
			panic("Data loaders is nil")
		}
		loader := loaders.GetLoader(p, asso, p.Args, model)
		thunk := loader.Load(p.Context, NewKey(source[consts.ID].(uint64)))
		return func() (interface{}, error) {
			data, err := thunk()
			if err != nil {
				return nil, handleError(err)
			}

			var retValue interface{}
			if data == nil {
				if asso.IsArray() {
					retValue = []map[string]interface{}{}
				} else {
					retValue = nil
				}
			} else {
				retValue = data
			}
			return retValue, nil
		}, nil
	}
}
