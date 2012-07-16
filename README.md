#  jfu - Go support for jQuery File Upload plugin

[Golang](http://golang.org) + [MongoDB](http://mongodb.org/) +
[Memcached](http://memcached.org) backend for the [jQuery File
Upload](http://blueimp.github.com/jQuery-File-Upload/) plugin.

All interaction with MongoDB is done thru a [simple
interface](http://go.pkgdoc.org/github.com/jmcvetta/jfu#DataStore), so adding
support for other datastores should be pretty easy.


# Documentation

See [pkgdoc](http://go.pkgdoc.org/github.com/jmcvetta/jfu) for automatic documentation.


# Example

See [jfu-example](https://github.com/jmcvetta/jfu-example) for a simple demo app.


# License

Released under the [WTFPL v2](http://sam.zoy.org/wtfpl/COPYING).

Derived from Sebastian Tschan's MIT-licensed example code, included with the
jQuery File Upload plugin, for a Google AppEngine-based Go backend.
