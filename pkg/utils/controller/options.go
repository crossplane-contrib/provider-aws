package controller

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
)

// Options allows to override controller.Options for specific controllers.
type Options struct {
	defaultOptions controller.Options
	specific       map[string]OptionsOverride
}

// OptionsOverride allows to override specific controller.Options properties.
type OptionsOverride struct {
	PollInterval            *time.Duration
	MaxConcurrentReconciles *int
}

func (override OptionsOverride) applyTo(options *controller.Options) {
	if override.PollInterval != nil {
		options.PollInterval = *override.PollInterval
	}

	if override.MaxConcurrentReconciles != nil {
		options.MaxConcurrentReconciles = *override.MaxConcurrentReconciles
	}
}

func NewOptions(defaultOptions controller.Options) Options {
	return Options{
		defaultOptions: defaultOptions,
		specific:       map[string]OptionsOverride{},
	}
}

// AddOverrides adds overrides for specific controllers from the provided map
// which is similar to ConfigMap data.
// Key format is "<scope>.<property>". Properties without scope or with "default" scope
// owerride default values.
func (options *Options) AddOverrides(values map[string]string) error {
	for key, value := range values {
		if err := options.addOverride(key, value); err != nil {
			return fmt.Errorf("failed to add override for %s: %w", key, err)
		}
	}
	return nil
}

func (options *Options) addOverride(key, value string) error {
	propSeparatorIdx := strings.LastIndex(key, ".")
	propName := key
	scope := "default"
	if propSeparatorIdx != -1 {
		propName = key[propSeparatorIdx+1:]
		scope = key[:propSeparatorIdx]
	}
	overrides := options.specific[scope]

	switch propName {
	case "pollInterval":
		if duration, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("failed to parse pollInterval value %s: %w", value, err)
		} else {
			overrides.PollInterval = &duration
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
		overrides.applyTo(&options.defaultOptions)
	} else {
		options.specific[scope] = overrides
	}
	return nil
}

// Default returns default controller.Options.
func (options Options) Default() controller.Options {
	return options.defaultOptions
}

// Get returns controller.Options for the specific controller.
func (options Options) Get(name string) controller.Options {
	result := options.defaultOptions
	if override, ok := options.specific[name]; ok {
		override.applyTo(&result)
	}
	return result
}
