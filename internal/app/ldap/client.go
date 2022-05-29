package ldap

import (
	"errors"
	"fmt"

	"github.com/go-ldap/ldap"
)

type Settings struct {
	BaseDn   string
	Host     string
}

type Client interface {
	GetUserInfoByUsername(login, password, username string) (*UserInfo, error)
}

type ldapClient struct {
	conn     *ldap.Conn
	baseDn   string
	host     string
}

type UserInfo struct {
	FullName string `json:"full_name"`
	Guid string `json:"guid"`
	SamAccountName string `json:"sam_account_name"`
	Email string `json:"email"`
}

func NewClient(settings Settings) (*ldapClient, error) {
	client := &ldapClient{
		baseDn:   settings.BaseDn,
		host:     settings.Host,
	}

	return client, nil
}

func (c *ldapClient) newConn(login, password string) error {
	conn, err := ldap.Dial("tcp", c.host)
	if err != nil {
		return fmt.Errorf("ldap.Dial host: %s, %w", c.host, err)
	}
	err = conn.Bind(login, password)
	if err != nil {
		conn.Close()
		return fmt.Errorf("conn.Bind host: %s, %w", c.host, err)
	}

	c.conn = conn

	return nil
}

func (c *ldapClient) GetUserInfoByUsername(login, password, username string) (*UserInfo, error) {
	err := c.newConn(login, password)
	if err != nil {
		return nil, fmt.Errorf("newConn: username %s: %w", username, err)
	}
	defer c.conn.Close()

	res, err := c.conn.Search(&ldap.SearchRequest{
		BaseDN: c.baseDn,
		Scope:      ldap.ScopeWholeSubtree,
		Filter:     fmt.Sprintf("(&(SamAccountName=%s))", username),
	})
	if err != nil {
		return nil, fmt.Errorf("conn.search: username %s: %w", username, err)
	}

	if len(res.Entries) != 1 {
		return nil, errors.New("user not found")
	}

	var info UserInfo
	for _, attribute := range res.Entries[0].Attributes {
		switch attribute.Name {
		case "displayName":
			info.FullName = attribute.Values[0]
		case "objectGUID":
			info.Guid = attribute.Values[0]
		case "sAMAccountName":
			info.SamAccountName = attribute.Values[0]
		case "mail":
			info.Email = attribute.Values[0]
		}
	}

	return &info, nil
}