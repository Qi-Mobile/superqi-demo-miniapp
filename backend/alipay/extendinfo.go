package alipay

import "encoding/json"

type ExtendInfoData struct {
	PaymentRequestID string                 `json:"paymentRequestId,omitempty"`
	UserID           string                 `json:"userId,omitempty"`
	OrderID          string                 `json:"orderId,omitempty"`
	ProductID        string                 `json:"productId,omitempty"`
	Quantity         int                    `json:"quantity,omitempty"`
	Timestamp        int64                  `json:"timestamp,omitempty"`
	CustomData       map[string]interface{} `json:"customData,omitempty"`
}

func BuildExtendInfo(data ExtendInfoData) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func ParseExtendInfo(extendInfo string) (ExtendInfoData, error) {
	var data ExtendInfoData
	if extendInfo == "" {
		return data, nil
	}

	err := json.Unmarshal([]byte(extendInfo), &data)
	return data, err
}

func BuildExtendInfoMap(dataMap map[string]interface{}) (string, error) {
	jsonBytes, err := json.Marshal(dataMap)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func ParseExtendInfoMap(extendInfo string) (map[string]interface{}, error) {
	var dataMap map[string]interface{}
	if extendInfo == "" {
		return dataMap, nil
	}

	err := json.Unmarshal([]byte(extendInfo), &dataMap)
	return dataMap, err
}
