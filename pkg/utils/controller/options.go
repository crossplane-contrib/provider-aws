package controller

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/controller"

	"github.com/crossplane-contrib/provider-aws/pkg/monitor"
)

// OptionsSet allows to override Options for specific controllers.
type OptionsSet struct {
	defaultOptions Options
	specific       map[string]OptionsOverride
}

type Options struct {
	controller.Options
	PollIntervalJitter time.Duration
	Monitor            monitor.Monitor
}

// OptionsOverride allows to override specific Options properties.
type OptionsOverride struct {
	PollInterval            *time.Duration
	PollIntervalJitter      *time.Duration
	MaxConcurrentReconciles *int
}

func (override OptionsOverride) applyTo(options *Options) {
	if override.PollInterval != nil {
		options.PollInterval = *override.PollInterval
	}

	if override.PollIntervalJitter != nil {
		options.PollIntervalJitter = *override.PollIntervalJitter
	}

	if override.MaxConcurrentReconciles != nil {
		options.MaxConcurrentReconciles = *override.MaxConcurrentReconciles
	}
}

func NewOptionsSet(defaultOptions Options) OptionsSet {
	return OptionsSet{
		defaultOptions: defaultOptions,
		specific:       map[string]OptionsOverride{},
	}
}

// AddOverrides adds overrides for specific controllers from the provided map
// which is similar to ConfigMap data.
// Key format is "<scope>.<property>". Properties without scope or with "default" scope
// owerride default values.
func (set *OptionsSet) AddOverrides(values map[string]string) error {
	for key, value := range values {
		if err := set.addOverride(key, value); err != nil {
			return fmt.Errorf("failed to add override for %s: %w", key, err)
		}
	}
	return nil
}

func (set *OptionsSet) addOverride(key, value string) error {
	propSeparatorIdx := strings.LastIndex(key, ".")
	propName := key
	scope := "default"
	if propSeparatorIdx != -1 {
		propName = key[propSeparatorIdx+1:]
		scope = key[:propSeparatorIdx]
	}
	overrides := set.specific[scope]

	switch propName {
	case "pollInterval":
		if duration, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("failed to parse pollInterval value %s: %w", value, err)
		} else {
			overrides.PollInterval = &duration
		}
	case "pollIntervalJitter":
		if duration, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("failed to parse pollIntervalJitter value %s: %w", value, err)
		} else {
			overrides.PollIntervalJitter = &duration
		}
	case "maxConcurrentReconciles":
		if maxConcurrentReconciles, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("failed to parse maxConcurrentReconciles value %s: %w", value, err)
		} else {
			overrides.MaxConcurrentReconciles = &maxConcurrentReconciles
		}
	default:
		return fmt.Errorf("unknown override property %s", propName)
	}

	if scope == "default" {
		overrides.applyTo(&set.defaultOptions)
	} else {
		set.specific[scope] = overrides
	}
	return nil
}

// Default returns default Options.
func (set OptionsSet) Default() Options {
	return set.defaultOptions
}

// Get returns Options for the specific controller.
func (set OptionsSet) Get(name string) Options {
	result := set.defaultOptions
	if override, ok := set.specific[name]; ok {
		override.applyTo(&result)
	}
	return result
}
