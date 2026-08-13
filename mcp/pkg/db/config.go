package db

import (
	"errors"
)

var ErrNoHostDefined = errors.New("no host defined")
var ErrNoUserDefined = errors.New("no user defined")
var ErrNoDatabaseNameDefined = errors.New("no database name defined")
var ErrNoPortDefined = errors.New("no port defined")
var ErrNoDatabaseFileDefined = errors.New("no database file defined")

type Config struct {
	DSN             string      `yaml:"dsn" json:"dsn"`
	Host            string      `yaml:"host" json:"host"`
	User            string      `yaml:"user" json:"user"`
	Password        string      `yaml:"password" json:"password"`
	DatabaseName    string      `yaml:"database_name" json:"database_name"`
	Port            int         `yaml:"port" json:"port"`
	ParseTime       bool        `yaml:"parse_time" json:"parse_time"`
	MultiStatements bool        `yaml:"multi_statements" json:"multi_statements"`
	TLS             string      `yaml:"tls" json:"tls"`
	TLSCAFile       string      `yaml:"tls_ca_file" json:"tls_ca_file"`
	DialectType     DialectType `yaml:"dialect_type" json:"dialect_type"`

	// DoltLite (embedded) configuration.
	// Path is the filesystem path of the DoltLite database file.
	Path string `yaml:"path" json:"path"`
	// CommitName and CommitEmail set the Dolt commit author on every
	// connection. DoltLite stores author config per connection and does not
	// persist it.
	CommitName  string `yaml:"commit_name" json:"commit_name"`
	CommitEmail string `yaml:"commit_email" json:"commit_email"`

	// doltLiteDatabase is initialized by PrepareDatabase for a server-owned
	// DoltLite pool. Config is passed by value throughout the tool layer, so
	// copies retain the same pool without exposing runtime state in serialized
	// configuration.
	doltLiteDatabase *doltLiteDatabase
}

func (c *Config) Validate() error {
	if c.DSN != "" {
		return nil
	}
	if c.DialectType == DialectDoltLite {
		if c.Path == "" {
			return ErrNoDatabaseFileDefined
		}
		return nil
	}
	if c.Host == "" {
		return ErrNoHostDefined
	}
	if c.User == "" {
		return ErrNoUserDefined
	}
	if c.Port == 0 {
		return ErrNoPortDefined
	}
	return nil
}
