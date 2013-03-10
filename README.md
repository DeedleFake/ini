ini
===

This [Go][golang] package provides a simple parser for [INI][ini]-based files. It can be used to parse many INI-derivative files, including .desktop files and systemd unit files.

Installation
------------

To install, type

> go get -v github.com/DeedleFake/ini

Usage
-----

This package has one main type, Parser, which has one method, Next(). To initialize a new parser, use

> p := ini.NewParser(r)

where r is an io.Reader. Several settings may then be configured. For more information, see the [API reference][godoc].

To parse the next token from the reader, call

> t, err := p.Next()

The returned token may be one of several different \*Token types, and is always returned as a pointer.

For full documentation and examples, see the [API reference][godoc].

Authors
-------

 * [DeedleFake](https://www.github.com/DeedleFake)

[golang]: http://www.golang.org
[ini]: http://www.wikipedia.com/wiki/INI_file
[godoc]: http://www.godoc.org/github.com/DeedleFake/ini
