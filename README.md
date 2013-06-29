cjdmap
======

cjdmap is a utility that traceroutes CJDNS nodes and prints output in the Nmap XML format.

* Is it useful?
* Does RadialNet/Zenmap faithfully represent node topology?
* Will all this pinging anger your peers?

I don't know.

```Bash
$ cjdmap HOST [HOST...] > map.xml
$ nmapfe map.xml
```

I ripped off all the useful code from [inhies](https://github.com/inhies).

Install
-------
`$ go get github.com/3M3RY/cjdmap`
