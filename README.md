cjdmap
======
cjdmap is a utility that outpus the local CJDNS routing table in the 
Nmap XML format.

Screenshot http://urlcloud.net/uuhM

```Bash
$ cjdmap [-all] [HOST...] > map.xml
$ nmapfe map.xml
```

cjdmap assumes that you have already created a .cjdnsadmin file with 
cjdcmd. Or you could create one manually:
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
