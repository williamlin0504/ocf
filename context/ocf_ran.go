package context

import (
	"fmt"
	"free5gc/lib/ngap/ngapConvert"
	"free5gc/lib/ngap/ngapType"
	"free5gc/lib/openapi/models"
	"free5gc/src/ocf/logger"
	"net"
)

const (
	RanPresentGNbId   = 1
	RanPresentNgeNbId = 2
	RanPresentN3IwfId = 3
)

type OcfRan struct {
	RanPresent int
	RanId      *models.GlobalRanNodeId
	Name       string
	AnType     models.AccessType
	/* socket Connect*/
	Conn net.Conn
	/* Supported TA List */
	SupportedTAList []SupportedTAI

	/* RAN UE List */
	RanUeList []*RanUe // RanUeNgapId as key
}

type SupportedTAI struct {
	Tai        models.Tai
	SNssaiList []models.Snssai
}

func NewSupportedTAI() (tai SupportedTAI) {
	tai.SNssaiList = make([]models.Snssai, 0, MaxNumOfSlice)
	return
}

func (ran *OcfRan) Remove() {
	ran.RemoveAllUeInRan()
	OCF_Self().DeleteOcfRan(ran.Conn)
}

func (ran *OcfRan) NewRanUe(ranUeNgapID int64) (*RanUe, error) {
	ranUe := RanUe{}
	self := OCF_Self()
	amfUeNgapID, err := self.AllocateOcfUeNgapID()
	if err != nil {
		return nil, fmt.Errorf("Allocate OCF UE NGAP ID error: %+v", err)
	}
	ranUe.OcfUeNgapId = amfUeNgapID
	ranUe.RanUeNgapId = ranUeNgapID
	ranUe.Ran = ran

	ran.RanUeList = append(ran.RanUeList, &ranUe)
	self.RanUePool.Store(ranUe.OcfUeNgapId, &ranUe)
	return &ranUe, nil
}

func (ran *OcfRan) RemoveAllUeInRan() {
	for _, ranUe := range ran.RanUeList {
		if err := ranUe.Remove(); err != nil {
			logger.ContextLog.Errorf("Remove RanUe error: %v", err)
		}
	}
}

func (ran *OcfRan) RanUeFindByRanUeNgapID(ranUeNgapID int64) *RanUe {
	for _, ranUe := range ran.RanUeList {
		if ranUe.RanUeNgapId == ranUeNgapID {
			return ranUe
		}
	}
	return nil
}

func (ran *OcfRan) SetRanId(ranNodeId *ngapType.GlobalRANNodeID) {
	ranId := ngapConvert.RanIdToModels(*ranNodeId)
	ran.RanPresent = ranNodeId.Present
	ran.RanId = &ranId
	if ranNodeId.Present == ngapType.GlobalRANNodeIDPresentGlobalN3IWFID {
		ran.AnType = models.AccessType_NON_3_GPP_ACCESS
	} else {
		ran.AnType = models.AccessType__3_GPP_ACCESS
	}
}
