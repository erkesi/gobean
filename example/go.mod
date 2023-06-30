module github.com/erkesi/gobean/example

go 1.18

replace github.com/erkesi/gobean/gstreamings => ../gstreamings

replace github.com/erkesi/gobean/gstreams => ../gstreams

require (
	github.com/erkesi/gobean/gstreamings v0.0.0-00010101000000-000000000000
	github.com/erkesi/gobean/gstreams v0.0.0-00010101000000-000000000000
)

require (
	github.com/erkesi/gobean v1.1.14 // indirect
	github.com/golang/mock v1.6.0 // indirect
)
