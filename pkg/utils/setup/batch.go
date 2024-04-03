package setup

import (
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-aws/pkg/utils/controller"
)

// Batch is a helper for setting up multiple controllers.
type Batch struct {
	manager ctrl.Manager
	options controller.Options
	prefix  string
	fns     []func() error
}

func NewBatch(manager ctrl.Manager, options controller.Options, prefix string) *Batch {
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

// AddUnscoped adds a controller setup function to the batch using default options.
func (b *Batch) AddUnscoped(fn SetupControllerFn) {
	b.fns = append(b.fns, func() error {
		return fn(b.manager, b.options.Default())
	})
}

// AddProxy adds a controller setup function to the batch that takes a controller.Options.
func (b *Batch) AddProxy(fn func(ctrl.Manager, controller.Options) error) {
	b.fns = append(b.fns, func() error {
		return fn(b.manager, b.options)
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
