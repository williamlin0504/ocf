//+build debug

package util

import (
	"free5gc/lib/path_util"
)

var OcfLogPath = path_util.Gofree5gcPath("free5gc/amfsslkey.log")
var OcfPemPath = path_util.Gofree5gcPath("free5gc/support/TLS/_debug.pem")
var OcfKeyPath = path_util.Gofree5gcPath("free5gc/support/TLS/_debug.key")
var DefaultOcfConfigPath = path_util.Gofree5gcPath("free5gc/config/amfcfg.conf")
