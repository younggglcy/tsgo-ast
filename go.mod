module github.com/younggglcy/tsgo-ast

go 1.26

replace (
	github.com/microsoft/typescript-go => ./tsgolint/typescript-go
	github.com/microsoft/typescript-go/shim/ast => ./tsgolint/shim/ast
	github.com/microsoft/typescript-go/shim/core => ./tsgolint/shim/core
	github.com/microsoft/typescript-go/shim/parser => ./tsgolint/shim/parser
	github.com/microsoft/typescript-go/shim/tspath => ./tsgolint/shim/tspath
)

require (
	github.com/microsoft/typescript-go/shim/ast v0.0.0
	github.com/microsoft/typescript-go/shim/core v0.0.0-00010101000000-000000000000
	github.com/microsoft/typescript-go/shim/parser v0.0.0-00010101000000-000000000000
	github.com/microsoft/typescript-go/shim/tspath v0.0.0-00010101000000-000000000000
)

require (
	github.com/go-json-experiment/json v0.0.0-20260214004413-d219187c3433 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/microsoft/typescript-go v0.0.0-20260309214900-4a59cd78390d // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace github.com/microsoft/typescript-go/shim/scanner => ./tsgolint/shim/scanner
