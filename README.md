streamtools
===========

    import "github.com/thejerf/streamtools"

One of the superpowers of Go is that the `io.Reader` and `io.Writer`
interfaces were in the library from the beginning. This has been quite
successful in encouraging the ecosystem to have pervasive streaming
available.

Streams have a lot of advantages in general, such as being able to operate
on vast quantities of input while consuming only a window's worth of
RAM. This is especially useful for network servers, which also happens to
be Go's wheelhouse, so the match is natural.

However, the programming community has decades of experience with strings,
which are resources fully in memory and with random access. While this is
sometimes unavoidable, it is advantageous to try as hard as possible when
dealing with streams to avoid turning them into strings. Turning streams
into large strings eliminates the advantages of streams.

That is intrinsically a challenge because working natively with streams is
difficult. I've seen some attempts to port other string APIs into Go, but
it would be preferable to have native support for readers and writers.

The purpose of this library is to provide native stream support based on
`io.Reader`s and `io.Writer`s.

This native support can be broken into two main categories:

1. Functions for interacting with streams directly, without intervening
   strings being created (until necessary).
2. Decorators that can be wrapped around `io.Reader`s or `io.Writer`s to
   provide some additional inline functionality.

Operating on streams has a fundamentally different set of idioms versus
operating on strings. It is important to understand that you will need
fundamentally different approaches to things in a streaming environment.

Release Status
==============

This package is currently **100% EXPERIMENTAL**. As I write this the one
big function it has is even broken and needs a fundamental rewrite.

Strictness
==========

The `io.Reader` and `io.Writer` interfaces have no room for issuing
warnings. There is only errors, likely terminating the stream, and panics,
certainly causing a mess.

However, there are a lot of things within this package that require certain
constraints on the incoming values, such as having buffers large enough to
hold certain results. It isn't necessarily appropriate to error out on
everything. The things that require such things will have a Strictness
parameter on them. Strict means that violations of the preconditions will
result in an `error` being returned, rather than values. Lax means that the
code will do its best to continue on.

There is a package-level `DefaultStrictness`; this should probably only be
set in your code. There is also a `LaxWarning` function that 

Semantic Versioning
===================

I expect this library to grow and develop over time. A true 1.0 may take
some time. The way I will work this is that the primary version number is
going to 0 probably for quite a while, but during development all API
changes will result in a change in the second version number. The changelog
will indicate what the change is.

I especially expect that it will take some development to reveal what the
naming scheme should be.

Features
========

* TBD.

Changelog
=========

* v0.0.4:
  * Boundary code converted into a state machine and should be correct now.
* v0.0.3:
  * Add io.Closer support to the BoundaryAtomicString. Start thinking about
    how to handle the variety of types that may be involved.
* v0.0.2:
  * The broken boundary code remains, code for consuming a reader until a
    byte or one of a set of bytes is added.
* v0.0.1: Initial release.
  * Broken code to make it possible to read an io.Reader stream for a
    specific string, and get a guarantee that that string will be
    atomically yielded as a .Read result if read.
