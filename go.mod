module github.com/code-ready/machine-driver-vf

go 1.16

require (
	github.com/Code-Hex/vz v0.0.5-0.20211218053248-d70a0533bf8e
	github.com/code-ready/crc v1.31.2
	github.com/code-ready/machine v0.0.0-20210616065635-eff475d32b9a
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/rs/xid v1.3.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	golang.org/x/text v0.3.7 // indirect
	inet.af/tcpproxy v0.0.0-20210824174053-2e577fef49e2
)

replace (
	github.com/apcera/gssapi => github.com/openshift/gssapi v0.0.0-20161010215902-5fb4217df13b
	github.com/containers/image => github.com/openshift/containers-image v0.0.0-20190130162819-76de87591e9d
	github.com/segmentio/analytics-go v3.2.0+incompatible => github.com/segmentio/analytics-go v1.2.1-0.20201110202747-0566e489c7b9
)

replace github.com/code-ready/machine => github.com/cfergeau/machine v0.0.0-20220118125514-4c7e58b2647a
