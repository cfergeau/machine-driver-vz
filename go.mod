module github.com/code-ready/machine-driver-vz

go 1.16

require (
	github.com/Code-Hex/vz v0.0.5-0.20210915161651-eccba26108a4
	github.com/code-ready/crc v1.31.2
	github.com/code-ready/machine v0.0.0-20210616065635-eff475d32b9a
	github.com/creack/pty v1.1.11 // indirect
	github.com/h2non/filetype v1.1.2-0.20210602110014-3305bbb7ac7b // indirect
	github.com/jinzhu/copier v0.3.2 // indirect
	github.com/klauspost/compress v1.13.4 // indirect
	github.com/kr/pty v1.1.8
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/pkg/term v1.1.0
	github.com/rs/xid v1.3.0 // indirect
	golang.org/x/sys v0.0.0-20210823070655-63515b42dcdf
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	golang.org/x/text v0.3.7 // indirect
)

replace (
	github.com/apcera/gssapi => github.com/openshift/gssapi v0.0.0-20161010215902-5fb4217df13b
	github.com/containers/image => github.com/openshift/containers-image v0.0.0-20190130162819-76de87591e9d
	github.com/segmentio/analytics-go v3.2.0+incompatible => github.com/segmentio/analytics-go v1.2.1-0.20201110202747-0566e489c7b9
)
