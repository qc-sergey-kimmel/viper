// Package viper provide viper bundle.
package viper

import (
	"os"
	"strings"

	"github.com/gozix/glue"
	"github.com/sarulabs/di"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

const (
	// BundleName is default definition name.
	BundleName = "viper"

	// DefConfigFlag is config persistent flag name.
	DefConfigFlag = "cli.persistent_flags.config"
)

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
	return builder.Add(
		di.Def{
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

				var cmd *cobra.Command
				if err = registry.Fill("cli.cmd", &cmd); err != nil {
					return nil, err
				}

				var configFile string
				if configFile, err = cmd.Flags().GetString("config"); err != nil {
					return nil, err
				}

				if len(configFile) > 0 {
					b.viper.SetConfigFile(configFile)
				}

				return b.viper, b.viper.ReadInConfig()
			},
		},
		di.Def{
			Name: DefConfigFlag,
			Tags: []di.Tag{{
				Name: glue.TagRootPersistentFlags,
			}},
			Build: func(_ di.Container) (i interface{}, e error) {
				var flagSet = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
				flagSet.StringP("config", "c", "", "config file")

				return flagSet, nil
			},
		},
	)
}

// apply implements Option.
func (f optionFunc) apply(bundle *Bundle) {
	f(bundle)
}
