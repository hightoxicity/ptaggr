// Package ptaggr implements a ptaggr plugin for CoreDNS
package ptaggr

import (
	"net"

	"golang.org/x/net/context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/dnsutil"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/plugin/pkg/nonwriter"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("ptaggr")

// Ptaggr plugin allows an extra set of upstreams be specified which will be used
// to serve an aggregated answer of all answers retrieved near those queried upstreams.
type Ptaggr struct {
	Next        plugin.Handler
	rules       []rule
	original    bool // At least one rule has "original" flag
	handlers    []HandlerWithCallbacks
	privateOnly bool
}

type rule struct {
	original bool
	handler  HandlerWithCallbacks
}

// HandlerWithCallbacks interface is made for handling the requests
type HandlerWithCallbacks interface {
	plugin.Handler
	OnStartup() error
	OnShutdown() error
}

// New initializes Alternate plugin
func New() (f *Ptaggr) {
	return &Ptaggr{}
}

// ServeDNS implements the plugin.Handler interface.
func (f Ptaggr) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	// If ptaggr has original option set for any rule then copy original request to use it instead of changed
	var originalRequest *dns.Msg
	if f.original {
		originalRequest = r.Copy()
	}
	nw := nonwriter.New(w)
	rcode, err := plugin.NextOrFailure(f.Name(), f.Next, ctx, nw, r)

	//By default the rulesIndex is equal rcode, so in such way we handle the case
	//when rcode is SERVFAIL and nw.Msg is nil, otherwise we use nw.Msg.Rcode
	//because, for example, for the following cases like NXDOMAIN, REFUSED the rcode is 0 (returned by forward)
	//A forward doesn't return 0 only in case SERVFAIL
	/*
		rulesIndex := rcode
		if nw.Msg != nil {
			rulesIndex = nw.Msg.Rcode
		}
	*/

	if state.QType() == dns.TypePTR && state.QClass() == dns.ClassINET {
		doProcess := true

		if f.privateOnly {
			ipStr := dnsutil.ExtractAddressFromReverse(state.Name())
			ip := net.ParseIP(ipStr)
			if ip.To4() == nil || !ip.IsPrivate() {
				doProcess = false
			}
		}

		if doProcess {
			var newR dns.Msg

			if len(nw.Msg.Answer) > 0 {
				newR = *nw.Msg
			}

			for _, ru := range f.rules {
				nwh := nonwriter.New(w)
				if ru.original && originalRequest != nil {
					ru.handler.ServeDNS(ctx, nwh, originalRequest)
				}
				_, err = ru.handler.ServeDNS(ctx, nwh, r)

				if err == nil {
					if len(nwh.Msg.Answer) > 0 {
						if len(newR.Answer) == 0 {
							newR = *nwh.Msg
						} else {
							for _, newRR := range nwh.Msg.Answer {
								if !isRRPresent(newRR, newR.Answer) {
									newR.Answer = append(newR.Answer, newRR)
								}
							}
						}
					}
				}
			}

			if len(newR.Answer) > 0 {
				w.WriteMsg(&newR)
				return newR.Rcode, err
			}
		}
	}

	if nw.Msg != nil {
		w.WriteMsg(nw.Msg)
	}
	return rcode, err
}

// Name implements the Handler interface.
func (f Ptaggr) Name() string { return "ptaggr" }

func isRRPresent(searchRR dns.RR, intoRRs []dns.RR) bool {
	present := false
	for _, rr := range intoRRs {
		if dns.IsDuplicate(rr, searchRR) {
			present = true
			break
		}
	}
	return present
}
