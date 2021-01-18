package callback

import (
	"context"
	"free5gc/lib/openapi/Namf_Communication"
	"free5gc/lib/openapi/models"
	amf_context "free5gc/src/ocf/context"
	"free5gc/src/ocf/logger"
	"reflect"
)

func SendOcfStatusChangeNotify(amfStatus string, guamiList []models.Guami) {
	amfSelf := amf_context.OCF_Self()

	amfSelf.OCFStatusSubscriptions.Range(func(key, value interface{}) bool {
		subscriptionData := value.(models.SubscriptionData)

		configuration := Namf_Communication.NewConfiguration()
		client := Namf_Communication.NewAPIClient(configuration)
		amfStatusNotification := models.OcfStatusChangeNotification{}
		var amfStatusInfo = models.OcfStatusInfo{}

		for _, guami := range guamiList {
			for _, subGumi := range subscriptionData.GuamiList {
				if reflect.DeepEqual(guami, subGumi) {
					//OCF status is available
					amfStatusInfo.GuamiList = append(amfStatusInfo.GuamiList, guami)
				}
			}
		}

		amfStatusInfo = models.OcfStatusInfo{
			StatusChange:     (models.StatusChange)(amfStatus),
			TargetOcfRemoval: "",
			TargetOcfFailure: "",
		}

		amfStatusNotification.OcfStatusInfoList = append(amfStatusNotification.OcfStatusInfoList, amfStatusInfo)
		uri := subscriptionData.OcfStatusUri

		logger.ProducerLog.Infof("[OCF] Send Ocf Status Change Notify to %s", uri)
		httpResponse, err := client.OcfStatusChangeCallbackDocumentApiServiceCallbackDocumentApi.
			OcfStatusChangeNotify(context.Background(), uri, amfStatusNotification)
		if err != nil {
			if httpResponse == nil {
				HttpLog.Errorln(err.Error())
			} else if err.Error() != httpResponse.Status {
				HttpLog.Errorln(err.Error())
			}
		}
		return true
	})
}
