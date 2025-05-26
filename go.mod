module gemserve

go 1.24.3

require (
	git.antanst.com/antanst/logging v0.0.0
	git.antanst.com/antanst/xerrors v0.0.0
	github.com/gabriel-vasile/mimetype v1.4.8
	github.com/matoous/go-nanoid/v2 v2.1.0
)

replace git.antanst.com/antanst/xerrors => ../xerrors

replace git.antanst.com/antanst/logging => ../logging

require golang.org/x/net v0.33.0 // indirect
