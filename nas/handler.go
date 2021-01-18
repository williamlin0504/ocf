package nas

import (
	"free5gc/src/ocf/context"
	"free5gc/src/ocf/logger"
	"free5gc/src/ocf/nas/nas_security"
)

func HandleNAS(ue *context.RanUe, procedureCode int64, nasPdu []byte) {
	amfSelf := context.OCF_Self()

	if ue == nil {
		logger.NasLog.Error("RanUe is nil")
		return
	}

	if nasPdu == nil {
		logger.NasLog.Error("nasPdu is nil")
		return
	}

	if ue.OcfUe == nil {
		ue.OcfUe = amfSelf.NewOcfUe("")
		ue.OcfUe.AttachRanUe(ue)
	}

	msg, err := nas_security.Decode(ue.OcfUe, ue.Ran.AnType, nasPdu)
	if err != nil {
		logger.NasLog.Errorln(err)
		return
	}

	if err := Dispatch(ue.OcfUe, ue.Ran.AnType, procedureCode, msg); err != nil {
		logger.NgapLog.Errorf("Handle NAS Error: %v", err)
	}
}
