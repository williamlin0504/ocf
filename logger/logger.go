package logger

import (
	"os"
	"time"

	formatter "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"

	"free5gc/lib/logger_conf"
	"free5gc/lib/logger_util"
)

var log *logrus.Logger
var AppLog *logrus.Entry
var InitLog *logrus.Entry
var ContextLog *logrus.Entry
var NgapLog *logrus.Entry
var HandlerLog *logrus.Entry
var HttpLog *logrus.Entry
var GmmLog *logrus.Entry
var MtLog *logrus.Entry
var ProducerLog *logrus.Entry
var LocationLog *logrus.Entry
var CommLog *logrus.Entry
var CallbackLog *logrus.Entry
var UtilLog *logrus.Entry
var NasLog *logrus.Entry
var ConsumerLog *logrus.Entry
var EeLog *logrus.Entry
var GinLog *logrus.Entry

func init() {
	log = logrus.New()
	log.SetReportCaller(false)

	log.Formatter = &formatter.Formatter{
		TimestampFormat: time.RFC3339,
		TrimMessages:    true,
		NoFieldsSpace:   true,
		HideKeys:        true,
		FieldsOrder:     []string{"component", "category"},
	}

	free5gcLogHook, err := logger_util.NewFileHook(logger_conf.Free5gcLogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err == nil {
		log.Hooks.Add(free5gcLogHook)
	}

	selfLogHook, err := logger_util.NewFileHook(logger_conf.NfLogDir+"ocf.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err == nil {
		log.Hooks.Add(selfLogHook)
	}

	AppLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "App"})
	InitLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "Init"})
	ContextLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "Context"})
	NgapLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "NGAP"})
	HandlerLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "Handler"})
	HttpLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "HTTP"})
	GmmLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "Gmm"})
	MtLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "MT"})
	ProducerLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "Producer"})
	LocationLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "LocInfo"})
	CommLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "Comm"})
	CallbackLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "Callback"})
	UtilLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "Util"})
	NasLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "NAS"})
	ConsumerLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "Consumer"})
	EeLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "EventExposure"})
	GinLog = log.WithFields(logrus.Fields{"component": "OCF", "category": "GIN"})
}

func SetLogLevel(level logrus.Level) {
	log.SetLevel(level)
}

func SetReportCaller(bool bool) {
	log.SetReportCaller(bool)
}
