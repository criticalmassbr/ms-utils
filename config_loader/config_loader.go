package configloader

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

type ConfigLoader[C any] interface {
	// Prefix defaults to "APP_" if not set
	DefaultInit(prefix string, overrideYAMLPath string) ConfigLoader[C]
	Env(prefix string) ConfigLoader[C]
	DB(db *sql.DB) ConfigLoader[C]
	YAML(path string) ConfigLoader[C]
	Validate() ConfigLoader[C]
	Context() (*C, error)
}

type ConfigContext[C any] struct {
	Config *C
}

var (
	ErrConfigNotSet = errors.New("config not set")
	ErrDSNNotSet    = errors.New("dsn not set")
)

type config[C any] struct {
	k                     *koanf.Koanf
	prefix                string
	delimiter             string
	validate              *validator.Validate
	err                   error
	configurationDBConfig *configurationDBConfig
	context               ConfigContext[C]
}

type configurationDBConfig struct {
	Config struct {
		DB struct {
			URL string `koanf:"url" validate:"required"`
		} `koanf:"db" validate:"required"`
	} `koanf:"config" validate:"required"`
}

func New[C any](validate *validator.Validate) ConfigLoader[C] {
	return &config[C]{
		k:         koanf.New("."),
		prefix:    "APP_",
		delimiter: ".",
		validate:  validate,
	}
}

func (c *config[C]) DefaultInit(prefix string, overrideYAMLPath string) ConfigLoader[C] {
	if c.err != nil {
		return c
	}

	if prefix != "" {
		c.prefix = prefix
	}

	k := koanf.New(".")

	if err := k.Load(env.Provider(c.prefix, c.delimiter, c.vaultEnvTransform), nil); err != nil {
		c.err = err
		return c
	}

	if overrideYAMLPath != "" {
		if err := k.Load(file.Provider(overrideYAMLPath), yaml.Parser()); err != nil {
			c.err = err
			return c
		}
	}

	if err := k.Unmarshal("", &c.configurationDBConfig); err != nil {
		c.err = err
		return c
	}

	return c
}

func (c *config[C]) Validate() ConfigLoader[C] {
	if c.err != nil && c.context.Config == nil {
		return c
	}

	if err := c.validate.Struct(c.context.Config); err != nil {
		c.err = err
		return c
	}

	return c
}

func (c *config[C]) DB(db *sql.DB) ConfigLoader[C] {
	if c.err != nil {
		return c
	}

	if db == nil {
		if err := c.validate.Struct(c.configurationDBConfig); err != nil {
			return c
		}

		configurationDB, err := sql.Open("postgres", c.configurationDBConfig.Config.DB.URL)
		if err != nil {
			c.err = err
			return c
		}
		db = configurationDB
		defer db.Close()
	}

	rows, err := db.Query(`select e."key", e.value from environment_variables e
				inner join (select "key", max(e.created_at) as created_at from environment_variables e group by e."key") e2
				on e."key" = e2."key" and e.created_at = e2.created_at`)
	if err != nil {
		c.err = err
		return c
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			c.err = err
			return c
		}

		c.k.Load(
			rawbytes.Provider([]byte(fmt.Sprintf("%s%s=%s", c.prefix, key, value))),
			dotenv.ParserEnv(c.prefix, c.delimiter, c.configurationDBEnvTransform),
		)

		os.Setenv(key, value)
	}

	if err := rows.Err(); err != nil {
		c.err = err
		return c
	}

	if err := c.k.Unmarshal("", &c.context.Config); err != nil {
		c.err = err
		return c
	}

	return c
}

func (c *config[C]) YAML(path string) ConfigLoader[C] {
	if c.err != nil || path == "" {
		return c
	}

	if err := c.k.Load(file.Provider(path), yaml.Parser()); err != nil {
		c.err = err
		return c
	}

	if err := c.k.Unmarshal("", &c.context.Config); err != nil {
		c.err = err
		return c
	}

	return c
}

func (c *config[C]) Env(prefix string) ConfigLoader[C] {
	if c.err != nil {
		return c
	}

	if prefix != "" {
		c.prefix = prefix
	}

	if err := c.k.Load(env.Provider(c.prefix, c.delimiter, c.vaultEnvTransform), nil); err != nil {
		c.err = err
		return c
	}

	if err := c.k.Unmarshal("", &c.context.Config); err != nil {
		c.err = err
		return c
	}

	return c
}

func (c *config[C]) Context() (*C, error) {
	if c.err != nil {
		return nil, c.err
	}

	if c.context.Config == nil {
		c.err = ErrConfigNotSet
		return nil, c.err
	}

	return c.context.Config, nil
}

func (c *config[C]) vaultEnvTransform(s string) string {
	return strings.
		Replace(strings.
			ToLower(strings.TrimPrefix(s, c.prefix)),
			"_",
			".",
			-1,
		)
}

func (c *config[C]) configurationDBEnvTransform(s string) string {
	return strings.ToLower(strings.TrimPrefix(s, c.prefix))
}
