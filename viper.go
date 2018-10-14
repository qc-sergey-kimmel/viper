package viper

import (
	"strings"

	"github.com/gozix/glue"
	"github.com/sarulabs/di"
	"github.com/spf13/viper"
)

type (
	// Option interface.
	Option interface {
		apply(bundle *Bundle)
	}

	// Bundle implements the glue.Bundle interface.
	Bundle struct {
		viper *viper.Viper
	}

	// Viper is type alias of viper.Viper
	Viper = viper.Viper

	// optionFunc wraps a func so it satisfies the Option interface.
	optionFunc func(bundle *Bundle)
)

// BundleName is default definition name.
const BundleName = "viper"

// NewBundle create bundle instance.
func NewBundle() *Bundle {
	return NewBundleWithConfig(
		AutomaticEnv(),
		EnvPrefix("ENV"),
		EnvKeyReplacer(strings.NewReplacer(".", "_")),
		ConfigName("config"),
		ConfigType("json"),
	)
}

// NewBundleWithConfig create bundle instance with config.
func NewBundleWithConfig(options ...Option) *Bundle {
	var bundle = Bundle{
		viper: viper.New(),
	}

	for _, option := range options {
		option.apply(&bundle)
	}

	return &bundle
}

// AutomaticEnv option.
func AutomaticEnv() Option {
	return optionFunc(func(bundle *Bundle) {
		bundle.viper.AutomaticEnv()
	})
}

// EnvPrefix option.
func EnvPrefix(value string) Option {
	return optionFunc(func(bundle *Bundle) {
		bundle.viper.SetEnvPrefix(value)
	})
}

// EnvKeyReplacer option.
func EnvKeyReplacer(value *strings.Replacer) Option {
	return optionFunc(func(bundle *Bundle) {
		bundle.viper.SetEnvKeyReplacer(value)
	})
}

// ConfigFile option.
func ConfigFile(value string) Option {
	return optionFunc(func(bundle *Bundle) {
		bundle.viper.SetConfigFile(value)
	})
}

// ConfigName option.
func ConfigName(value string) Option {
	return optionFunc(func(bundle *Bundle) {
		bundle.viper.SetConfigName(value)
	})
}

// ConfigPath option.
func ConfigPath(value string) Option {
	return optionFunc(func(bundle *Bundle) {
		bundle.viper.AddConfigPath(value)
	})
}

// ConfigType option.
func ConfigType(value string) Option {
	return optionFunc(func(bundle *Bundle) {
		bundle.viper.SetConfigType(value)
	})
}

// Name implements the glue.Bundle interface.
func (b *Bundle) Name() string {
	return BundleName
}

// Build implements the glue.Bundle interface.
func (b *Bundle) Build(builder *di.Builder) error {
	return builder.Add(di.Def{
		Name: BundleName,
		Build: func(ctn di.Container) (_ interface{}, err error) {
			var registry glue.Registry
			if err = ctn.Fill(glue.DefRegistry, &registry); err != nil {
				return nil, err
			}

			var path string
			if err = registry.Fill("app.path", &path); err != nil {
				return nil, err
			}

			b.viper.AddConfigPath(path)

			return b.viper, b.viper.ReadInConfig()
		},
	})
}

// apply implements Option.
func (f optionFunc) apply(bundle *Bundle) {
	f(bundle)
}
