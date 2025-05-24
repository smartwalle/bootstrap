module github.com/smartwalle/bootstrap/example

go 1.23.0

require (
	golang.org/x/sync v0.13.0
	github.com/smartwalle/bootstrap v0.0.0
)

replace (
	github.com/smartwalle/bootstrap => ../
)
