package setup

import (
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-aws/pkg/utils/controller"
)

// SetupControllerFn is a delegate to initialize a controller with provider-aws Options.
type SetupControllerFn func(ctrl.Manager, controller.Options) error //nolint:golint

// Batch is a helper for setting up multiple controllers.
type Batch struct {
	manager ctrl.Manager
	options controller.OptionsSet
	prefix  string
	fns     []func() error
}

func NewBatch(manager ctrl.Manager, options controller.OptionsSet, prefix string) *Batch {
	if prefix != "" && !strings.HasSuffix(prefix, ".") {
		prefix += "."
	}
	return &Batch{
		manager: manager,
		options: options,
		prefix:  prefix,
	}
}

// Add adds a controller setup function to the batch, scoping its options with a given name.
func (b *Batch) Add(name string, fn SetupControllerFn) {
	b.fns = append(b.fns, func() error {
		return fn(b.manager, b.options.Get(b.prefix+name))
	})
}

// Add adds a controller setup function to the batch, scoping its options with a given name.
// Setup function takes crossplane Options.
func (b *Batch) AddXp(name string, fn SetupControllerFnXp) {
	b.fns = append(b.fns, func() error {
		return fn(b.manager, b.options.Get(b.prefix+name).Options)
	})
}

// AddProxy adds a controller setup function to the batch.
// Setup function takes a controller.OptionsSet.
func (b *Batch) AddProxy(fn func(ctrl.Manager, controller.OptionsSet) error) {
	b.fns = append(b.fns, func() error {
		return fn(b.manager, b.options)
	})
}

// AddProxyXp adds a controller setup function to the batch using default crossplane options.
// Setup function takes crossplane Options.
func (b *Batch) AddProxyXp(fn SetupControllerFnXp) {
	b.fns = append(b.fns, func() error {
		return fn(b.manager, b.options.Default().Options)
	})
}

// Run runs all controller setup functions in the batch.
func (b *Batch) Run() error {
	for _, fn := range b.fns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
