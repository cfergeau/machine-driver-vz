module github.com/code-ready/machine-driver-vf

go 1.16

require (
	github.com/Code-Hex/vz v0.0.5-0.20211218053248-d70a0533bf8e
	github.com/code-ready/machine v0.0.0-20210616065635-eff475d32b9a
	github.com/docker/go-units v0.4.0
	github.com/rs/xid v1.3.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.3.0
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9
	inet.af/tcpproxy v0.0.0-20210824174053-2e577fef49e2
)

replace (
	github.com/apcera/gssapi => github.com/openshift/gssapi v0.0.0-20161010215902-5fb4217df13b
	github.com/containers/image => github.com/openshift/containers-image v0.0.0-20190130162819-76de87591e9d
	github.com/segmentio/analytics-go v3.2.0+incompatible => github.com/segmentio/analytics-go v1.2.1-0.20201110202747-0566e489c7b9
)

replace github.com/code-ready/machine => github.com/cfergeau/machine v0.0.0-20220118125514-4c7e58b2647a
