package storage

const (
	ExternalServiceTypeUnspecified ExternalServiceType = 0
	ExternalServiceTypeYandex ExternalServiceType = 1
	ExternalServiceTypeVkontakte ExternalServiceType = 2
	ExternalServiceTypeGoogle ExternalServiceType = 3
)

type ExternalServiceType int8

type Link struct {
	EmployeeId string `db:"employee_id"`
	ExternalServiceId string `db:"external_service_id"`
	ExternalServiceTypeId ExternalServiceType `db:"external_service_type_id"`
}