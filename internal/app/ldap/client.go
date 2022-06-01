package ldap

import (
	"errors"
	"fmt"

	"github.com/go-ldap/ldap"
)

type Settings struct {
	BaseDn   string
	Host     string
	Username string
	Password string
}

type Client interface {
	GetUserInfoByUsername(username string) (*UserInfo, error)
	GetEmployeeNumberByEmail(email string) (string, error)
}

type ldapClient struct {
	conn     *ldap.Conn
	baseDn   string
	host     string
	username    string
	password string
}

type UserInfo struct {
	FullName string `json:"full_name"`
	EmployeeNumber string `json:"employee_number"`
	SamAccountName string `json:"sam_account_name"`
	Email string `json:"email"`
}

func NewClient(settings Settings) (*ldapClient, error) {
	client := &ldapClient{
		baseDn:   settings.BaseDn,
		host:     settings.Host,
		username: settings.Username,
		password: settings.Password,
	}

	return client, nil
}

func (c *ldapClient) newConn() error {
	conn, err := ldap.Dial("tcp", c.host)
	if err != nil {
		return fmt.Errorf("ldap.Dial host: %s, %w", c.host, err)
	}
	err = conn.Bind(c.username, c.password)
	if err != nil {
		conn.Close()
		return fmt.Errorf("conn.Bind host: %s, %w", c.host, err)
	}

	c.conn = conn

	return nil
}

func (c *ldapClient) GetUserInfoByUsername(username string) (*UserInfo, error) {
	err := c.newConn()
	if err != nil {
		return nil, fmt.Errorf("newConn: %w", err)
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
	res.PrettyPrint(1)
	if len(res.Entries) != 1 {
		return nil, errors.New("user not found")
	}

	var info UserInfo
	for _, attribute := range res.Entries[0].Attributes {
		switch attribute.Name {
		case "displayName":
			info.FullName = attribute.Values[0]
		case "employeeNumber":
			info.EmployeeNumber = attribute.Values[0]
		case "sAMAccountName":
			info.SamAccountName = attribute.Values[0]
		case "mail":
			info.Email = attribute.Values[0]
		}
	}

	return &info, nil
}

func (c *ldapClient) GetEmployeeNumberByEmail(email string) (string, error) {
	err := c.newConn()
	if err != nil {
		return "", fmt.Errorf("newConn: %w", err)
	}
	defer c.conn.Close()

	res, err := c.conn.Search(&ldap.SearchRequest{
		BaseDN: c.baseDn,
		Scope:      ldap.ScopeWholeSubtree,
		Filter:     fmt.Sprintf("(&(mail=%s))", email),
	})
	if err != nil {
		return "", fmt.Errorf("conn.search: email %s: %w", email, err)
	}
	if len(res.Entries) != 1 {
		return "", nil
	}

	for _, attribute := range res.Entries[0].Attributes {
		if attribute.Name == "employeeNumber" {
			return attribute.Values[0], nil
		}
	}

	return "", nil
}