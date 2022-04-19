package ptraggr

import (
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/coredns/caddy"
)

func init() {
	caddy.RegisterPlugin("ptraggr", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	a := New()

	for c.Next() {
		// shift cursor past ptraggr
		if !c.Next() {
			return c.ArgErr()
		}

		var (
			original    bool
			privateOnly bool
			err         error
		)

		if original, err = getOriginal(c); err != nil {
			return err
		}

		if privateOnly, err = getPrivate(c); err != nil {
			return err
		}

		handlers, err := initForwards(c)

		if err != nil {
			return plugin.Error("ptraggr", err)
		}

		for _, hd := range handlers {
			a.handlers = append(a.handlers, hd)
			a.rules = append(a.rules, rule{original: original, handler: hd})
		}

		if original {
			a.original = true
		}

		if privateOnly {
			a.privateOnly = true
		}
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		a.Next = next
		return a
	})

	c.OnShutdown(func() error {
		for _, handler := range a.handlers {
			if err := handler.OnShutdown(); err != nil {
				return err
			}
		}
		return nil
	})

	return nil
}

const original = "original"
const privateOnly = "privateonly"

func getOriginal(c *caddy.Controller) (bool, error) {
	if c.Val() == original {
		// shift cursor past original
		if !c.Next() {
			return false, c.ArgErr()
		}
		return true, nil
	}

	return false, nil
}

func getPrivate(c *caddy.Controller) (bool, error) {
	if c.Val() == privateOnly {
		// shift cursor past private
		if !c.Next() {
			return false, c.ArgErr()
		}
		return true, nil
	}

	return false, nil
}
