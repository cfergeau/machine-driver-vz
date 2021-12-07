module github.com/code-ready/machine-driver-vz

go 1.16

require (
	github.com/Code-Hex/vz v0.0.5-0.20211213160915-f0b7da742b9f
	github.com/code-ready/crc v1.37.0
	github.com/code-ready/machine v0.0.0-20210616065635-eff475d32b9a
	github.com/kr/pty v1.1.8
	github.com/pkg/term v1.1.0
	github.com/rs/xid v1.3.0 // indirect
	golang.org/x/sys v0.0.0-20211210111614-af8b64212486
)

replace (
	github.com/apcera/gssapi => github.com/openshift/gssapi v0.0.0-20161010215902-5fb4217df13b
	github.com/containers/image => github.com/openshift/containers-image v0.0.0-20190130162819-76de87591e9d
	github.com/segmentio/analytics-go v3.2.0+incompatible => github.com/segmentio/analytics-go v1.2.1-0.20201110202747-0566e489c7b9
)
