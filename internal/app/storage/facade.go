package storage

import "context"

type Facade interface {
	CreateLink(ctx context.Context, link Link) error
	GetEmployeeId(ctx context.Context, externalServiceId string, externalServiceType ExternalServiceType) (string, error)
}

type facade struct {
	pg *PgStorage
}

func NewStorageFacade(pg *PgStorage) Facade {
	return &facade{pg: pg}
}

func (f *facade) CreateLink(ctx context.Context, link Link) error {
	return f.pg.CreateLink(ctx, link)
}

func (f *facade) GetEmployeeId(ctx context.Context, externalServiceId string, externalServiceType ExternalServiceType) (string, error) {
	return f.pg.GetEmployeeId(ctx, externalServiceId, externalServiceType)
}

