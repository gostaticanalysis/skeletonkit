module github.com/gostaticanalysis/skeletonkit

go 1.23.3

require (
	github.com/josharian/txtarfs v0.0.0-20240408113805-5dc76b8fe6bf
	github.com/tenntenn/golden v0.2.0
	golang.org/x/mod v0.22.0
	golang.org/x/tools v0.27.0
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/josharian/mapfs v0.0.0-20210615234106-095c008854e6 // indirect
	golang.org/x/sync v0.9.0 // indirect
)

retract (
	v0.3.0
	v0.1.0
)
