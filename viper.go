// Package viper provide viper bundle.
package viper

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/gozix/glue/v2"
	"github.com/sarulabs/di/v2"
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

var (
	// ErrUndefinedAppPath is error, triggered when app.path is undefined in current context.
	ErrUndefinedAppPath = errors.New("app.path is undefined")

	// ErrUndefinedCliCmd is error, triggered when cli.cmd is undefined in current context.
	ErrUndefinedCliCmd = errors.New("cli.cmd is undefined")
)

const (
	// BundleName is default definition name.
	BundleName = "viper"

	// DefConfigFlag is config persistent flag name.
	DefConfigFlag = "cli.persistent_flags.config"
)

// NewBundle create bundle instance.
func NewBundle(options ...Option) *Bundle {
	var opts = []Option{
		AutomaticEnv(),
		EnvPrefix("ENV"),
		EnvKeyReplacer(strings.NewReplacer(".", "_")),
		ConfigName("config"),
		ConfigType("json"),
	}

	opts = append(opts, options...)

	return NewBundleWithConfig(opts...)
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
				var ctx context.Context
				if err = ctn.Fill(glue.DefContext, &ctx); err != nil {
					return nil, err
				}

				var path, ok = ctx.Value("app.path").(string)
				if !ok {
					return nil, ErrUndefinedAppPath
				}

				b.viper.AddConfigPath(path)

				var cmd *cobra.Command
				if cmd, ok = ctx.Value("cli.cmd").(*cobra.Command); !ok {
					return nil, ErrUndefinedCliCmd
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
