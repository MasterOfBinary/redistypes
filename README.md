Redis data types in Go
======================

[![Build Status](https://travis-ci.org/MasterOfBinary/redistypes.svg?branch=master)](https://travis-ci.org/MasterOfBinary/redistypes)
[![Coverage Status](https://coveralls.io/repos/github/MasterOfBinary/redistypes/badge.svg?branch=master)](https://coveralls.io/github/MasterOfBinary/redistypes?branch=master)
[![GoDoc](https://godoc.org/github.com/MasterOfBinary/redistypes?status.svg)](https://godoc.org/github.com/MasterOfBinary/redistypes)

This is a very thin wrapper around redigo that provides a convenient way to use Redis's data types in Go.

Features
--------

Go implementations of the following data types in Redis:

1. List (partial)
2. HyperLogLog

More to come!

Documentation
-------------

See the [GoDocs](https://godoc.org/github.com/MasterOfBinary/redistypes).

Installation
------------

To download, run

    go get github.com/MasterOfBinary/redistypes

Redistypes requires the following dependencies:

* https://github.com/golang/groupcache/singleflight
* https://github.com/garyburd/redigo/redis

Example
-------

For a full, runnable example, see https://github.com/MasterOfBinary/goredistypes.

License
-------

Redistypes is provided under the MIT licence. See `LICENSE` for more details.
