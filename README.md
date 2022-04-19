# ptaggr

## Name

Plugin *Ptaggr* is able to aggregate PTR request answers of several upstream servers.

## Description

The *ptaggr* plugin allows an extra set of upstreams be specified which will be used
to serve an aggregated answer of all answers retrieved near those queried upstreams. The *ptaggr* plugin utilizes the *forward* plugin (<https://coredns.io/plugins/forward>) to query the specified upstreams.

> The *ptaggr* plugin supports only DNS protocol and random policy w/o additional *forward* parameters, so following directives will fail:

```
. {
    forward . 8.8.8.8
    ptaggr . tls://192.168.1.1:853 {
        policy sequential
    }
}
```

As the name suggests, the purpose of the *ptaggr* is to allow several primary dns servers to give their own answer for a reverse request, you may need this when you have for example several subzones managed by several DNS authoritative servers that do not share any root in term of DNS tree.

## Syntax

```
{
    ptaggr [original] [privateonly] . DNS_RESOLVERS
}
```

* **original** is optional flag. If it is set then ptaggr uses original request instead of potentially changed by other plugins
* **privateonly** is optional flag. If it is set then aggregation mechanism happen only for requested reverse IPv4 requests from RFC 1918 private ranges
* **DNS_RESOLVERS** accepts dns resolvers list. Requests will be forked near all those upstreams.

## Building CoreDNS with ptaggr

When building CoreDNS with this plugin, _ptaggr_ should be positioned **before** _forward_ in `/plugin.cfg`.

## Examples

### Ptaggr aggregated reverse answer with 3 upstreams

The following specifies that all requests are forwarded to 8.8.8.8, 1.1.1.1 and 208.67.222.222.

```
. {
    forward in-addr.arpa 8.8.8.8
    ptaggr in-addr.arpa 1.1.1.1:53 208.67.222.222:53
    log
}

```
