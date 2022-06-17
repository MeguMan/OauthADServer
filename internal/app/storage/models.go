package storage

const (
	ExternalServiceTypeUnspecified ExternalServiceType = 0
	ExternalServiceTypeYandex ExternalServiceType = 1
	ExternalServiceTypeVkontakte ExternalServiceType = 2
	ExternalServiceTypeGoogle ExternalServiceType = 3
	ExternalServiceTypeGithub ExternalServiceType = 4
	ExternalServiceTypeMail ExternalServiceType = 5
	ExternalServiceTypeDiscord ExternalServiceType = 6
	ExternalServiceTypeFacebook ExternalServiceType = 7

	LoginStatusOk LoginStatus = "ok"
	LoginStatusFail LoginStatus = "fail"
)

type ExternalServiceType int8
type LoginStatus string

type Link struct {
	EmployeeId string `db:"employee_id"`
	ExternalServiceId string `db:"external_service_id"`
	ExternalServiceTypeId ExternalServiceType `db:"external_service_type_id"`
}