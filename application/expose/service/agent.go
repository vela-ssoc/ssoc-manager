package service

//type Agent struct {
//	db  repository.Database
//	log *slog.Logger
//}
//
//func NewAgent(db repository.Database, log *slog.Logger) *Agent {
//	return &Agent{
//		db:  db,
//		log: log,
//	}
//}
//
//func (agt *Agent) Get(ctx context.Context, id int64) (*model.Minion, error) {
//	tbl := agt.qry.Minion
//	dao := tbl.WithContext(ctx)
//
//	return dao.Where(tbl.ID.Eq(id)).First()
//}
