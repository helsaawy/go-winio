//go:build windows

package etw

import "github.com/Microsoft/go-winio/pkg/etw"

type Opt func(*exporter) error

// WithNewETWProvider registers a new ETW provider for the exporter to use.
// The provider will be closed when the exporter is shutdown.
func WithNewETWProvider(n string) Opt {
	return func(e *exporter) error {
		provider, err := etw.NewProvider(n, nil)
		if err != nil {
			return err
		}

		e.p = provider
		e.closeProvider = true
		return nil
	}
}

// WithExistingETWProvider configures the exporter to use an existing ETW provider.
// The provider will not be closed when the exporter is shutdown.
func WithExistingETWProvider(p *etw.Provider) Opt {
	return func(e *exporter) error {
		e.p = p
		e.closeProvider = false
		return nil
	}
}
