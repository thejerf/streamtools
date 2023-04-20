/*
Package advstreamtools provides access to more generic streaming tools,
based on the io.Reader and io.Writer interface, but generalized beyond
byte streams to streams of anything.

The algorithms exposed by the top level of the streamtools archive are
generally trivially generalizable beyond byte streams. Byte streams
are a particularly tricky case to deal with because whatever entities
a consumer of a byte stream may want to deal with is very likely to
cross .Read calls. Streams of "other things" are more often something
that you can deal with one "thing" at a time, so a simple loop to
consume them one at a time and reading whene necessary is all that is
necessary.

However, every once in a while that is not the case, and if you
encounter such a case, there's no reason for me not to expose the
generalized algorithms for your utility and pleasure.

So this package exposes streaming tools against GeneralReader and
GeneralWriter, which are the obvious generic extensions of the
famous io.Reader and io.Writer interfaces.
*/
package advstreamtools
