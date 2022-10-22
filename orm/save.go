package orm

import (
	"log"

	"rxdrag.com/entify/consts"
	"rxdrag.com/entify/db/dialect"
	"rxdrag.com/entify/model/data"
)

func (s *Session) SaveOne(instance *data.Instance) (interface{}, error) {
	if instance.IsInsert() {
		return s.insertOne(instance)
	} else {
		return s.updateOne(instance)
	}
}

func (s *Session) insertOne(instance *data.Instance) (interface{}, error) {
	sqlBuilder := dialect.GetSQLBuilder()
	saveStr := sqlBuilder.BuildInsertSQL(instance.Fields, instance.Table())
	values := makeFieldValues(instance.Fields)
	result, err := s.Dbx.Exec(saveStr, values...)
	if err != nil {
		log.Println("Insert data failed:", err.Error())
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Println("LastInsertId failed:", err.Error())
		return nil, err
	}
	for _, asso := range instance.Associations {
		err = s.saveAssociation(asso, uint64(id))
		if err != nil {
			log.Println("Save reference failed:", err.Error())
			return nil, err
		}
	}

	savedObject := s.QueryOneEntityById(instance.Entity, id)

	//affectedRows, err := result.RowsAffected()
	if err != nil {
		log.Println("RowsAffected failed:", err.Error())
		return nil, err
	}

	return savedObject, nil
}

func (con *Session) updateOne(instance *data.Instance) (interface{}, error) {

	sqlBuilder := dialect.GetSQLBuilder()

	saveStr := sqlBuilder.BuildUpdateSQL(instance.Id, instance.Fields, instance.Table())
	values := makeFieldValues(instance.Fields)
	log.Println(saveStr)
	_, err := con.Dbx.Exec(saveStr, values...)
	if err != nil {
		log.Println("Update data failed:", err.Error())
		return nil, err
	}

	for _, ref := range instance.Associations {
		con.saveAssociation(ref, instance.Id)
	}

	savedObject := con.QueryOneEntityById(instance.Entity, instance.Id)

	return savedObject, nil
}

func newAssociationPovit(r *data.AssociationRef, ownerId uint64, tarId uint64) *data.AssociationPovit {
	sourceId := ownerId
	targetId := tarId

	if !r.IsSource() {
		sourceId = targetId
		targetId = ownerId
	}

	return data.NewAssociationPovit(r, sourceId, targetId)

}

func (con *Session) saveAssociationInstance(ins *data.Instance) (interface{}, error) {
	targetData := InsanceData{consts.ID: ins.Id}

	saved, err := con.SaveOne(ins)
	if err != nil {
		return nil, err
	}
	targetData = saved.(InsanceData)

	return targetData, nil
}
func (con *Session) saveAssociation(r *data.AssociationRef, ownerId uint64) error {

	for _, ins := range r.Deleted {
		if r.Cascade() {
			con.DeleteInstance(ins)
		} else {
			povit := newAssociationPovit(r, ownerId, ins.Id)
			con.DeleteAssociationPovit(povit)
		}
	}

	for _, ins := range r.Added {
		targetData, err := con.saveAssociationInstance(ins)

		if err != nil {
			panic("Save Association error:" + err.Error())
		} else {
			if savedIns, ok := targetData.(InsanceData); ok {
				tarId := savedIns[consts.ID].(uint64)
				relationInstance := newAssociationPovit(r, ownerId, tarId)
				con.SaveAssociationPovit(relationInstance)
			} else {
				panic("Save Association error")
			}
		}

	}

	for _, ins := range r.Updated {
		// if ins.Id == 0 {
		// 	panic("Can not add new instance when update")
		// }
		targetData, err := con.saveAssociationInstance(ins)
		if err != nil {
			panic("Save Association error:" + err.Error())
		} else {
			if savedIns, ok := targetData.(InsanceData); ok {
				tarId := savedIns[consts.ID].(uint64)
				relationInstance := newAssociationPovit(r, ownerId, tarId)

				con.SaveAssociationPovit(relationInstance)
			} else {
				panic("Save Association error")
			}
		}
	}

	synced := r.Synced
	if len(synced) == 0 {
		return nil
	}

	con.clearAssociation(r, ownerId)

	for _, ins := range synced {
		targetId := ins.Id
		if !ins.IsEmperty() {
			targetData, err := con.saveAssociationInstance(ins)
			if err != nil {
				panic("Save Association error:" + err.Error())
			} else {
				if savedIns, ok := targetData.(InsanceData); ok {
					targetId = savedIns[consts.ID].(uint64)
				} else {
					panic("Save Association error")
				}
			}
		}
		relationInstance := newAssociationPovit(r, ownerId, targetId)
		con.SaveAssociationPovit(relationInstance)
	}

	return nil
}

func (con *Session) SaveAssociationPovit(povit *data.AssociationPovit) {
	sqlBuilder := dialect.GetSQLBuilder()
	sql := sqlBuilder.BuildQueryPovitSQL(povit)
	rows, err := con.Dbx.Query(sql)
	defer rows.Close()
	if err != nil {
		panic(err.Error())
	}
	if !rows.Next() {
		sql = sqlBuilder.BuildInsertPovitSQL(povit)
		_, err := con.Dbx.Exec(sql)
		if err != nil {
			panic(err.Error())
		}
	}
}
