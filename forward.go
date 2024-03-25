package ptaggr

import (
	"fmt"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/forward"
	"github.com/coredns/coredns/plugin/pkg/parse"
	"github.com/coredns/coredns/plugin/pkg/proxy"
	"github.com/coredns/coredns/plugin/pkg/transport"
)

const defaultExpire = 10 * time.Second

func initForwards(c *caddy.Controller) ([]*forward.Forward, error) {
	var f []*forward.Forward

	to := c.RemainingArgs()

	if len(to) == 0 {
		return f, c.ArgErr()
	}

	toHosts, err := parse.HostPortOrFile(to...)
	if err != nil {
		return f, err
	}

	for c.NextBlock() {
		return f, fmt.Errorf("additional parameters not allowed")
	}

	for _, host := range toHosts {
		trans, h := parse.Transport(host)
		if trans != transport.DNS {
			return f, fmt.Errorf("only dns transport allowed")
		}

		p := proxy.NewProxy("ptaggr", h, trans)
		p.SetExpire(defaultExpire)
		fw := forward.New()
		fw.SetProxy(p)
		f = append(f, fw)
	}

	return f, nil
}
