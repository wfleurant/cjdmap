cjdmap
======
cjdmap is a utility that outpus the local CJDNS routing table in the 
Nmap XML format.

Zenmap(nmapfe) will not plot multiple routes for single host, so rather
than attempt to find any and all routes, cjdmap outputs the route returned
by the CJDNS RouterModule_lookup function.

I think these routes are accurate, but make for sufficient estimates.

Screenshot http://urlcloud.net/uuhM

```Bash
$ cjdmap HOST1 HOST2 > map.xml
$ nmapfe map.xml
```

cjdmap assumes that you have a ~/.cjdnsadmin file with 
cjdcmd. This file is shared with cjdcmd and other utilities.
The format is as follows:
```JSON
{
    "addr": "127.0.0.1",
    "port": 11234,
    "password": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "config": "/etc/cjdroute.conf"
}
```

I ripped off all the useful code from [inhies](https://github.com/inhies).

Todo
----
Reverse hostname resolution.

Install
-------
`$ go get github.com/3M3RY/cjdmap`
