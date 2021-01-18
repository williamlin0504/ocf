package context

import (
	"fmt"
	"free5gc/lib/idgenerator"
	"free5gc/lib/openapi/models"
	"free5gc/src/ocf/logger"
	"math"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var amfContext = OCFContext{}
var tmsiGenerator *idgenerator.IDGenerator = nil
var amfUeNGAPIDGenerator *idgenerator.IDGenerator = nil
var amfStatusSubscriptionIDGenerator *idgenerator.IDGenerator = nil

func init() {
	OCF_Self().LadnPool = make(map[string]*LADN)
	OCF_Self().EventSubscriptionIDGenerator = idgenerator.NewGenerator(1, math.MaxInt32)
	OCF_Self().Name = "ocf"
	OCF_Self().UriScheme = models.UriScheme_HTTPS
	OCF_Self().RelativeCapacity = 0xff
	OCF_Self().ServedGuamiList = make([]models.Guami, 0, MaxNumOfServedGuamiList)
	OCF_Self().PlmnSupportList = make([]PlmnSupportItem, 0, MaxNumOfPLMNs)
	OCF_Self().NfService = make(map[models.ServiceName]models.NfService)
	OCF_Self().NetworkName.Full = "free5GC"
	tmsiGenerator = idgenerator.NewGenerator(1, math.MaxInt32)
	amfStatusSubscriptionIDGenerator = idgenerator.NewGenerator(1, math.MaxInt32)
	amfUeNGAPIDGenerator = idgenerator.NewGenerator(1, MaxValueOfOcfUeNgapId)
}

type OCFContext struct {
	EventSubscriptionIDGenerator    *idgenerator.IDGenerator
	EventSubscriptions              sync.Map
	UePool                          sync.Map         // map[supi]*OcfUe
	RanUePool                       sync.Map         // map[OcfUeNgapID]*RanUe
	OcfRanPool                      sync.Map         // map[net.Conn]*OcfRan
	LadnPool                        map[string]*LADN // dnn as key
	SupportTaiLists                 []models.Tai
	ServedGuamiList                 []models.Guami
	PlmnSupportList                 []PlmnSupportItem
	RelativeCapacity                int64
	NfId                            string
	Name                            string
	NfService                       map[models.ServiceName]models.NfService // nfservice that ocf support
	UriScheme                       models.UriScheme
	BindingIPv4                     string
	SBIPort                         int
	RegisterIPv4                    string
	HttpIPv6Address                 string
	TNLWeightFactor                 int64
	SupportDnnLists                 []string
	OCFStatusSubscriptions          sync.Map // map[subscriptionID]models.SubscriptionData
	NrfUri                          string
	SecurityAlgorithm               SecurityAlgorithm
	NetworkName                     NetworkName
	NgapIpList                      []string // NGAP Server IP
	T3502Value                      int      // unit is second
	T3512Value                      int      // unit is second
	Non3gppDeregistrationTimerValue int      // unit is second
}

type OCFContextEventSubscription struct {
	IsAnyUe           bool
	IsGroupUe         bool
	UeSupiList        []string
	Expiry            *time.Time
	EventSubscription models.OcfEventSubscription
}

type PlmnSupportItem struct {
	PlmnId     models.PlmnId   `yaml:"plmnId"`
	SNssaiList []models.Snssai `yaml:"snssaiList,omitempty"`
}

type NetworkName struct {
	Full  string `yaml:"full"`
	Short string `yaml:"short,omitempty"`
}

type SecurityAlgorithm struct {
	IntegrityOrder []uint8 // slice of security.AlgIntegrityXXX
	CipheringOrder []uint8 // slice of security.AlgCipheringXXX
}

func NewPlmnSupportItem() (item PlmnSupportItem) {
	item.SNssaiList = make([]models.Snssai, 0, MaxNumOfSlice)
	return
}

func (context *OCFContext) TmsiAllocate() int32 {
	tmsi, err := tmsiGenerator.Allocate()
	if err != nil {
		logger.ContextLog.Errorf("Allocate TMSI error: %+v", err)
		return -1
	}
	return int32(tmsi)
}

func (context *OCFContext) AllocateOcfUeNgapID() (int64, error) {
	return amfUeNGAPIDGenerator.Allocate()
}

func (context *OCFContext) AllocateGutiToUe(ue *OcfUe) {
	servedGuami := context.ServedGuamiList[0]
	ue.Tmsi = context.TmsiAllocate()

	plmnID := servedGuami.PlmnId.Mcc + servedGuami.PlmnId.Mnc
	tmsiStr := fmt.Sprintf("%08x", ue.Tmsi)
	ue.Guti = plmnID + servedGuami.OcfId + tmsiStr
}

func (context *OCFContext) AllocateRegistrationArea(ue *OcfUe, anType models.AccessType) {

	// clear the previous registration area if need
	if len(ue.RegistrationArea[anType]) > 0 {
		ue.RegistrationArea[anType] = nil
	}

	// allocate a new tai list as a registration area to ue
	// TODO: algorithm to choose TAI list
	for _, supportTai := range context.SupportTaiLists {
		if reflect.DeepEqual(supportTai, ue.Tai) {
			ue.RegistrationArea[anType] = append(ue.RegistrationArea[anType], supportTai)
			break
		}
	}
}

func (context *OCFContext) NewOCFStatusSubscription(subscriptionData models.SubscriptionData) (subscriptionID string) {
	id, err := amfStatusSubscriptionIDGenerator.Allocate()
	if err != nil {
		logger.ContextLog.Errorf("Allocate subscriptionID error: %+v", err)
		return ""
	}

	subscriptionID = strconv.Itoa(int(id))
	context.OCFStatusSubscriptions.Store(subscriptionID, subscriptionData)
	return
}

// Return Value: (subscriptionData *models.SubScriptionData, ok bool)
func (context *OCFContext) FindOCFStatusSubscription(subscriptionID string) (*models.SubscriptionData, bool) {
	if value, ok := context.OCFStatusSubscriptions.Load(subscriptionID); ok {
		subscriptionData := value.(models.SubscriptionData)
		return &subscriptionData, ok
	} else {
		return nil, false
	}
}

func (context *OCFContext) DeleteOCFStatusSubscription(subscriptionID string) {
	context.OCFStatusSubscriptions.Delete(subscriptionID)
	if id, err := strconv.ParseInt(subscriptionID, 10, 64); err != nil {
		logger.ContextLog.Error(err)
	} else {
		amfStatusSubscriptionIDGenerator.FreeID(id)
	}
}

func (context *OCFContext) NewEventSubscription(subscriptionID string, subscription *OCFContextEventSubscription) {
	context.EventSubscriptions.Store(subscriptionID, subscription)
}

func (context *OCFContext) FindEventSubscription(subscriptionID string) (*OCFContextEventSubscription, bool) {
	if value, ok := context.EventSubscriptions.Load(subscriptionID); ok {
		return value.(*OCFContextEventSubscription), ok
	} else {
		return nil, false
	}
}
func (context *OCFContext) DeleteEventSubscription(subscriptionID string) {
	context.EventSubscriptions.Delete(subscriptionID)
	if id, err := strconv.ParseInt(subscriptionID, 10, 32); err != nil {
		logger.ContextLog.Error(err)
	} else {
		context.EventSubscriptionIDGenerator.FreeID(id)
	}
}

func (context *OCFContext) AddOcfUeToUePool(ue *OcfUe, supi string) {
	if len(supi) == 0 {
		logger.ContextLog.Errorf("Supi is nil")
	}
	ue.Supi = supi
	context.UePool.Store(ue.Supi, ue)
}

func (context *OCFContext) NewOcfUe(supi string) *OcfUe {
	ue := OcfUe{}
	ue.init()

	if supi != "" {
		context.AddOcfUeToUePool(&ue, supi)
	}

	context.AllocateGutiToUe(&ue)

	return &ue
}

func (context *OCFContext) OcfUeFindByUeContextID(ueContextID string) (*OcfUe, bool) {
	if strings.HasPrefix(ueContextID, "imsi") {
		return context.OcfUeFindBySupi(ueContextID)
	}
	if strings.HasPrefix(ueContextID, "imei") {
		return context.OcfUeFindByPei(ueContextID)
	}
	if strings.HasPrefix(ueContextID, "5g-guti") {
		guti := ueContextID[strings.LastIndex(ueContextID, "-")+1:]
		return context.OcfUeFindByGuti(guti)
	}
	return nil, false
}

func (context *OCFContext) OcfUeFindBySupi(supi string) (ue *OcfUe, ok bool) {
	if value, loadOk := context.UePool.Load(supi); loadOk {
		ue = value.(*OcfUe)
		ok = loadOk
	}
	return
}

func (context *OCFContext) OcfUeFindByPei(pei string) (ue *OcfUe, ok bool) {
	context.UePool.Range(func(key, value interface{}) bool {
		candidate := value.(*OcfUe)
		if ok = (candidate.Pei == pei); ok {
			ue = candidate
			return false
		}
		return true
	})
	return
}

func (context *OCFContext) NewOcfRan(conn net.Conn) *OcfRan {
	ran := OcfRan{}
	ran.SupportedTAList = make([]SupportedTAI, 0, MaxNumOfTAI*MaxNumOfBroadcastPLMNs)
	ran.Conn = conn
	context.OcfRanPool.Store(conn, &ran)
	return &ran
}

// use net.Conn to find RAN context, return *OcfRan and ok bit
func (context *OCFContext) OcfRanFindByConn(conn net.Conn) (*OcfRan, bool) {
	if value, ok := context.OcfRanPool.Load(conn); ok {
		return value.(*OcfRan), ok
	}
	return nil, false
}

// use ranNodeID to find RAN context, return *OcfRan and ok bit
func (context *OCFContext) OcfRanFindByRanID(ranNodeID models.GlobalRanNodeId) (*OcfRan, bool) {
	var ran *OcfRan
	var ok bool
	context.OcfRanPool.Range(func(key, value interface{}) bool {
		amfRan := value.(*OcfRan)
		switch amfRan.RanPresent {
		case RanPresentGNbId:
			logger.ContextLog.Infof("aaa: %+v\n", amfRan.RanId.GNbId)
			if amfRan.RanId.GNbId.GNBValue == ranNodeID.GNbId.GNBValue {
				ran = amfRan
				ok = true
				return false
			}
		case RanPresentNgeNbId:
			if amfRan.RanId.NgeNbId == ranNodeID.NgeNbId {
				ran = amfRan
				ok = true
				return false
			}
		case RanPresentN3IwfId:
			if amfRan.RanId.N3IwfId == ranNodeID.N3IwfId {
				ran = amfRan
				ok = true
				return false
			}
		}
		return true
	})
	return ran, ok
}

func (context *OCFContext) DeleteOcfRan(conn net.Conn) {
	context.OcfRanPool.Delete(conn)
}

func (context *OCFContext) InSupportDnnList(targetDnn string) bool {
	for _, dnn := range context.SupportDnnLists {
		if dnn == targetDnn {
			return true
		}
	}
	return false
}

func (context *OCFContext) OcfUeFindByGuti(guti string) (ue *OcfUe, ok bool) {
	context.UePool.Range(func(key, value interface{}) bool {
		candidate := value.(*OcfUe)
		if ok = (candidate.Guti == guti); ok {
			ue = candidate
			return false
		}
		return true
	})
	return
}

func (context *OCFContext) OcfUeFindByPolicyAssociationID(polAssoId string) (ue *OcfUe, ok bool) {
	context.UePool.Range(func(key, value interface{}) bool {
		candidate := value.(*OcfUe)
		if ok = (candidate.PolicyAssociationId == polAssoId); ok {
			ue = candidate
			return false
		}
		return true
	})
	return
}

func (context *OCFContext) RanUeFindByOcfUeNgapID(amfUeNgapID int64) *RanUe {
	if value, ok := context.RanUePool.Load(amfUeNgapID); ok {
		return value.(*RanUe)
	} else {
		return nil
	}
}

func (context *OCFContext) GetIPv4Uri() string {
	return fmt.Sprintf("%s://%s:%d", context.UriScheme, context.RegisterIPv4, context.SBIPort)
}

func (context *OCFContext) InitNFService(serivceName []string, version string) {
	tmpVersion := strings.Split(version, ".")
	versionUri := "v" + tmpVersion[0]
	for index, nameString := range serivceName {
		name := models.ServiceName(nameString)
		context.NfService[name] = models.NfService{
			ServiceInstanceId: strconv.Itoa(index),
			ServiceName:       name,
			Versions: &[]models.NfServiceVersion{
				{
					ApiFullVersion:  version,
					ApiVersionInUri: versionUri,
				},
			},
			Scheme:          context.UriScheme,
			NfServiceStatus: models.NfServiceStatus_REGISTERED,
			ApiPrefix:       context.GetIPv4Uri(),
			IpEndPoints: &[]models.IpEndPoint{
				{
					Ipv4Address: context.RegisterIPv4,
					Transport:   models.TransportProtocol_TCP,
					Port:        int32(context.SBIPort),
				},
			},
		}
	}
}

// Reset OCF Context
func (context *OCFContext) Reset() {
	context.OcfRanPool.Range(func(key, value interface{}) bool {
		context.UePool.Delete(key)
		return true
	})
	for key := range context.LadnPool {
		delete(context.LadnPool, key)
	}
	context.RanUePool.Range(func(key, value interface{}) bool {
		context.RanUePool.Delete(key)
		return true
	})
	context.UePool.Range(func(key, value interface{}) bool {
		context.UePool.Delete(key)
		return true
	})
	context.EventSubscriptions.Range(func(key, value interface{}) bool {
		context.DeleteEventSubscription(key.(string))
		return true
	})
	for key := range context.NfService {
		delete(context.NfService, key)
	}
	context.SupportTaiLists = context.SupportTaiLists[:0]
	context.PlmnSupportList = context.PlmnSupportList[:0]
	context.ServedGuamiList = context.ServedGuamiList[:0]
	context.RelativeCapacity = 0xff
	context.NfId = ""
	context.UriScheme = models.UriScheme_HTTPS
	context.SBIPort = 0
	context.BindingIPv4 = ""
	context.RegisterIPv4 = ""
	context.HttpIPv6Address = ""
	context.Name = "ocf"
	context.NrfUri = ""
}

// Create new OCF context
func OCF_Self() *OCFContext {
	return &amfContext
}
