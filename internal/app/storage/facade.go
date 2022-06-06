package storage

import "context"

type Facade interface {
	CreateLink(ctx context.Context, link Link) error
	GetEmployeeId(ctx context.Context, externalServiceId string, externalServiceType ExternalServiceType) (string, error)
	//CreateClient(ctx context.Context, client OauthClient) error
	//GetRedirectUriByClientId(ctx context.Context, clientId string) (string, error)
	//GetRedirectUriByClientIdAndSecret(ctx context.Context, client OauthClient) (string, error)
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

//func (f *facade) CreateClient(ctx context.Context, client OauthClient) error {
//	return f.pg.CreateClient(ctx, client)
//}
//
//func (f *facade) GetRedirectUriByClientId(ctx context.Context, clientId string) (string, error) {
//	return f.pg.GetRedirectUriByClientId(ctx, clientId)
//}
//
//func (f *facade) GetRedirectUriByClientIdAndSecret(ctx context.Context, client OauthClient) (string, error) {
//	return f.pg.GetRedirectUriByClientIdAndSecret(ctx, client)
//}


