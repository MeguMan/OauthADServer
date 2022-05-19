package ldap

import (
	"errors"
	"fmt"

	"github.com/go-ldap/ldap"
)

const (
	disableAccountCode = "514"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrManagerNotFound = errors.New("manager not found")
	ErrAccountDisable  = errors.New("account disable")
)

type Settings struct {
	Host     string
	Username string
	Password string
	BaseDn   string
}

type UserInfo struct {
	UserAccountControl string
}

type Client interface {
	GetUserInfoByUsername(username string) (*UserInfo, error)
}

type ldapClient struct {
	conn     *ldap.Conn
	baseDn   string
	username string
	password string
	host     string
}

func NewClient(settings Settings) (*ldapClient, error) {
	client := &ldapClient{
		baseDn:   settings.BaseDn,
		host:     settings.Host,
		password: settings.Password,
		username: settings.Username,
	}

	// пробуем подключиться к лдап серверу
	err := client.newConn()
	if err != nil {
		return nil, fmt.Errorf("connect to ldap client: %w", err)
	}
	client.conn.Close()

	return client, nil
}

func (c *ldapClient) newConn() error {
	conn, err := ldap.Dial("tcp", c.host)
	if err != nil {
		return fmt.Errorf("NewConn: ldap.Dial host: %s, %w", c.host, err)
	}
	err = conn.Bind(c.username, c.password)
	if err != nil {
		conn.Close()
		return fmt.Errorf("NewConn: conn.Bind host: %s, %w", c.host, err)
	}

	c.conn = conn

	return nil
}

func (c *ldapClient) GetUserInfoByUsername(username string) (*UserInfo, error) {
	err := c.newConn()
	if err != nil {
		return nil, fmt.Errorf("GetUserInfoByUsername: username %s: %w", username, err)
	}
	defer c.conn.Close()

	res, err := c.conn.Search(&ldap.SearchRequest{
		BaseDN:     c.baseDn,
		Scope:      ldap.ScopeWholeSubtree,
		Filter:     fmt.Sprintf("(&(SamAccountName=%s))", username),
		Attributes: []string{"userAccountControl"},
	})
	if err != nil {
		return nil, fmt.Errorf("GetUserInfoByUsername: username %s: %w", username, err)
	}
	if len(res.Entries) != 1 {
		return nil, ErrUserNotFound
	}

	for _, attribute := range res.Entries[0].Attributes {
		if attribute.Name == "userAccountControl" {
			if attribute.Values[0] == disableAccountCode {
				return nil, ErrAccountDisable
			}
		}
	}

	return &UserInfo{}, nil
}