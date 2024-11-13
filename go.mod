module github.com/gostaticanalysis/skeletonkit

go 1.23.3

require (
	github.com/josharian/txtarfs v0.0.0-20210615234325-77aca6df5bca
	github.com/tenntenn/golden v0.2.0
	golang.org/x/mod v0.4.2
	golang.org/x/tools v0.1.7
)

require (
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/josharian/mapfs v0.0.0-20210615234106-095c008854e6 // indirect
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

retract (
	v0.3.0
	v0.1.0
)
