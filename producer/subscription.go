package producer

import (
	"free5gc/lib/http_wrapper"
	"free5gc/lib/openapi/models"
	"free5gc/src/ocf/context"
	"free5gc/src/ocf/logger"
	"net/http"
	"reflect"
)

// TS 29.518 5.2.2.5.1
func HandleOCFStatusChangeSubscribeRequest(request *http_wrapper.Request) *http_wrapper.Response {
	logger.CommLog.Info("Handle OCF Status Change Subscribe Request")

	subscriptionDataReq := request.Body.(models.SubscriptionData)

	subscriptionDataRsp, locationHeader, problemDetails := OCFStatusChangeSubscribeProcedure(subscriptionDataReq)
	if problemDetails != nil {
		return http_wrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}

	headers := http.Header{
		"Location": {locationHeader},
	}
	return http_wrapper.NewResponse(http.StatusCreated, headers, subscriptionDataRsp)
}

func OCFStatusChangeSubscribeProcedure(subscriptionDataReq models.SubscriptionData) (
	subscriptionDataRsp models.SubscriptionData, locationHeader string, problemDetails *models.ProblemDetails) {
	amfSelf := context.OCF_Self()

	for _, guami := range subscriptionDataReq.GuamiList {
		for _, servedGumi := range amfSelf.ServedGuamiList {
			if reflect.DeepEqual(guami, servedGumi) {
				//OCF status is available
				subscriptionDataRsp.GuamiList = append(subscriptionDataRsp.GuamiList, guami)
			}
		}
	}

	if subscriptionDataRsp.GuamiList != nil {
		newSubscriptionID := amfSelf.NewOCFStatusSubscription(subscriptionDataReq)
		locationHeader = subscriptionDataReq.OcfStatusUri + "/" + newSubscriptionID
		logger.CommLog.Infof("new OCF Status Subscription[%s]", newSubscriptionID)
		return
	} else {
		problemDetails = &models.ProblemDetails{
			Status: http.StatusForbidden,
			Cause:  "UNSPECIFIED",
		}
		return
	}
}

// TS 29.518 5.2.2.5.2
func HandleOCFStatusChangeUnSubscribeRequest(request *http_wrapper.Request) *http_wrapper.Response {
	logger.CommLog.Info("Handle OCF Status Change UnSubscribe Request")

	subscriptionID := request.Params["subscriptionId"]

	problemDetails := OCFStatusChangeUnSubscribeProcedure(subscriptionID)
	if problemDetails != nil {
		return http_wrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	} else {
		return http_wrapper.NewResponse(http.StatusNoContent, nil, nil)
	}
}

func OCFStatusChangeUnSubscribeProcedure(subscriptionID string) (problemDetails *models.ProblemDetails) {
	amfSelf := context.OCF_Self()

	if _, ok := amfSelf.FindOCFStatusSubscription(subscriptionID); !ok {
		problemDetails = &models.ProblemDetails{
			Status: http.StatusNotFound,
			Cause:  "SUBSCRIPTION_NOT_FOUND",
		}
	} else {
		logger.CommLog.Debugf("Delete OCF status subscription[%s]", subscriptionID)
		amfSelf.DeleteOCFStatusSubscription(subscriptionID)
	}
	return
}

// TS 29.518 5.2.2.5.1.3
func HandleOCFStatusChangeSubscribeModify(request *http_wrapper.Request) *http_wrapper.Response {
	logger.CommLog.Info("Handle OCF Status Change Subscribe Modify Request")

	updateSubscriptionData := request.Body.(models.SubscriptionData)
	subscriptionID := request.Params["subscriptionId"]

	updatedSubscriptionData, problemDetails := OCFStatusChangeSubscribeModifyProcedure(subscriptionID,
		updateSubscriptionData)
	if problemDetails != nil {
		return http_wrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	} else {
		return http_wrapper.NewResponse(http.StatusAccepted, nil, updatedSubscriptionData)
	}
}

func OCFStatusChangeSubscribeModifyProcedure(subscriptionID string, subscriptionData models.SubscriptionData) (
	*models.SubscriptionData, *models.ProblemDetails) {
	amfSelf := context.OCF_Self()

	if currentSubscriptionData, ok := amfSelf.FindOCFStatusSubscription(subscriptionID); !ok {
		problemDetails := &models.ProblemDetails{
			Status: http.StatusForbidden,
			Cause:  "Forbidden",
		}
		return nil, problemDetails
	} else {
		logger.CommLog.Debugf("Modify OCF status subscription[%s]", subscriptionID)

		currentSubscriptionData.GuamiList = currentSubscriptionData.GuamiList[:0]

		currentSubscriptionData.GuamiList = append(currentSubscriptionData.GuamiList, subscriptionData.GuamiList...)
		currentSubscriptionData.OcfStatusUri = subscriptionData.OcfStatusUri

		amfSelf.OCFStatusSubscriptions.Store(subscriptionID, currentSubscriptionData)
		return currentSubscriptionData, nil
	}
}
