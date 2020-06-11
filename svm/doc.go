// Copyright (c) 2019-present Wasmer, Inc. and its affiliates.
// Copyright (c) 2018-present Spacemesh

// [SVM] is a smart contracts framework written in [Rust].
// Its runtime exposes a C API through [the `svm-runtime-c-api` crate][svm-runtime-c-api],
// which is also written in Rust but compiles to C-compatible shared
// libraries. C and C++ headers are also automatically built at
// compile-time. Build artifacts are checked into this package
// To allow direct and seamless import.
//
// Go provides [cgo] that enables the creation of Go packages that
// call C code. This package uses cgo to communicate with SVM through
// C.
//
// The `bridge.go` file contains a thin layer on top of the cgo
// generated code to get a more user-friendly API. It is the only place
// where data transit from Go to the SVM's C API. Any new features
// provided by SVM that needs to be exposed in this package must be
// “described” there.
//
// Schematically, the workflow looks like this:
//
//     +------------------------------+
//     |                              |
//     |  +------------------------+  |
//     |  |                        |  |
//     |  |          Go            |  |
//     |  |                        |  |
//     |  +----+--------------+----+  |
//     |       |              ^       |
//     |       v              |       |
//     |  +----+--------------+----+  |
//     |  |                        |  |
//     |  |       bridge.go        |  |   go-svm/svm package
//     |  |                        |  |
//     |  +----+--------------+----+  |
//     |       |              ^       |
//     |       v              |       |
//     |  +----+--------------+----+  |
//     |  |                        |  |
//     |  |          cgo           |  |
//     |  |                        |  |
//     |  +----+--------------+----+  |
//     |       |              ^       |
//     +-------|--------------|-------+
//     |       v              |       |
//     |  +----+--------------+----+  |
//     |  |                        |  |
//     |  |   svm-runtime-c-api    |  |   (shared library)
//     |  |                        |  |
//     |  +----+--------------+----+  |
//     |       |              ^       |
//     |       v              |       |
//     |  +----+--------------+----+  |
//     |  |                        |  |
//     |  |      svm-runtime       |  |
//     |  |                        |  |
//     |  +------------------------+  |
//     |                              |
//     +------------------------------+
//
// The cgo part is auto-generated by Go. It should be considered
// as an opaque black-box.
//
// Thanks to `bridge.go`, the rest of this package can talk to
// SVM as if it is almost regular Go code.
//
// [SVM]: https://github.com/spacemeshos/svm
// [Rust]: https://www.rust-lang.org/
// [svm-runtime-c-api]: https://github.com/spacemeshos/svm/tree/master/crates/svm-runtime-c-api
// [cgo]: https://golang.org/cmd/cgo/
package svm