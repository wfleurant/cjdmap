cjdmap
======
cjdmap is a utility that traceroutes CJDNS nodes and prints output in the Nmap XML format.
Screenshot http://urlcloud.net/uuhM

At this point it does not actually ping nodes to find latency, but simply maps out entries in the local routing table.
Pinging will happen when Nmap compatibility improves.

```Bash
$ cjdmap HOST [HOST...] > map.xml
$ nmapfe map.xml
```

I ripped off all the useful code from [inhies](https://github.com/inhies).

Todo
----
Dump the IPv4 address of UDPInterface peers.
Map entire local routing table.
Host resolution.
Pinging

Install
-------
`$ go get github.com/3M3RY/cjdmap`
