// Copyright 2018 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

// Package viper provide viper bundle.
package viper

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gozix/di"
	"github.com/gozix/glue/v3"
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

	// optionFunc wraps a func, so it satisfies the Option interface.
	optionFunc func(bundle *Bundle)
)

// ErrUndefinedAppPath is error, triggered when app.path is undefined in current context.
var ErrUndefinedAppPath = errors.New("app.path is undefined")

const (
	// BundleName is default definition name.
	BundleName = "viper"

	// tagViperFlagSet is tag marks bundle flag set.
	tagViperFlagSet = "viper.flag_set"
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
func (b *Bundle) Build(builder di.Builder) error {
	return builder.Apply(
		di.Provide(
			b.provideViper,
			di.Constraint(1, di.WithTags(tagViperFlagSet)),
		),
		di.Provide(b.provideFlagSet, glue.AsPersistentFlags(), di.Tags{{
			Name: tagViperFlagSet,
		}}),
	)
}

func (b *Bundle) provideViper(ctx context.Context, flagSet *pflag.FlagSet) (_ *viper.Viper, err error) {
	var path, ok = ctx.Value("app.path").(string)
	if !ok {
		return nil, ErrUndefinedAppPath
	}

	b.viper.AddConfigPath(path)

	var configFile string
	if configFile, err = flagSet.GetString("config"); err != nil {
		return nil, fmt.Errorf("unable to get config flag value : %w", err)
	}

	if len(configFile) > 0 {
		b.viper.SetConfigFile(configFile)
	}

	err = b.viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to read config file : '%s' : %w",
			configFile, err)
	}

	return b.viper, nil
}

func (b *Bundle) provideFlagSet() (*pflag.FlagSet, error) {
	var flagSet = pflag.NewFlagSet(BundleName, pflag.ContinueOnError)
	flagSet.StringP("config", "c", "", "config file")

	var err = flagSet.Parse(os.Args)
	if errors.Is(err, pflag.ErrHelp) {
		err = nil
	}

	return flagSet, err
}

// apply implements Option.
func (f optionFunc) apply(bundle *Bundle) {
	f(bundle)
}
