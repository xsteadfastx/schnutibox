//nolint:gochecknoglobals
package assets

import _ "embed"

//go:embed templates/schnutibox.yml.tmpl
var SchnutiboxConfig string

//go:embed files/schnutibox.service
var SchnutiboxService []byte

//go:embed templates/fstab.raspbian.tmpl
var FstabRaspbianTemplate string

//go:embed templates/mopidy.conf.tmpl
var MopidyConf string

//go:embed files/upmpdcli.conf
var UpmpdcliConf []byte

//go:embed files/ntp.service
var NtpService []byte

//go:embed templates/50-neuftech.rules.tmpl
var UDEVRules string
