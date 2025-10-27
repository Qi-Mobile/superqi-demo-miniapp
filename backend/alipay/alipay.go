package alipay

import (
	"encoding/json"
	"log"
)

func (client *Client) ApplyToken(authCode string) (ApplyTokenResponse, error) {
	const path = "/v1/authorizations/applyToken"
	params := map[string]string{
		"grantType": "AUTHORIZATION_CODE",
		"authCode":  authCode,
	}

	headers, err := client.buildHeaders("POST", path, params)
	if err != nil {
		return ApplyTokenResponse{}, err
	}

	response, err := client.sendRequest(path, "POST", headers, params)
	if err != nil {
		return ApplyTokenResponse{}, err
	}

	var body ApplyTokenResponse
	err = json.Unmarshal(response, &body)
	return body, err
}

func (client *Client) InquiryUserInfo(accessToken string) (InquiryUserInfoResponse, error) {
	const path = "/v1/users/inquiryUserInfo"
	params := map[string]string{
		"accessToken": accessToken,
	}

	headers, err := client.buildHeaders("POST", "/v1/users/inquiryUserInfo", params)
	if err != nil {
		return InquiryUserInfoResponse{}, err
	}

	response, err := client.sendRequest(path, "POST", headers, params)
	if err != nil {
		return InquiryUserInfoResponse{}, err
	}

	var body InquiryUserInfoResponse
	err = json.Unmarshal(response, &body)
	return body, err
}

func (client *Client) PrepareAuthorization(contractDescription string) (PrepareAuthorizationResponse, error) {
	const path = "/v1/authorizations/prepare"

	extendInfoMap := map[string]string{
		"language":     "en-US",
		"contractDesc": contractDescription,
	}

	extendInfoJSON, err := json.Marshal(extendInfoMap)
	if err != nil {
		return PrepareAuthorizationResponse{}, err
	}

	params := map[string]interface{}{
		"scopes":     "AGREEMENT_PAY",
		"extendInfo": string(extendInfoJSON),
	}

	log.Println("[Alipay Client] Preparing authorization request")
	log.Printf("[Alipay Client] Scopes: %s\n", params["scopes"])
	log.Printf("[Alipay Client] ExtendInfo: %s\n", params["extendInfo"])

	headers, err := client.buildHeaders("POST", path, params)
	if err != nil {
		log.Printf("[Alipay Client] ERROR: Failed to build headers: %v\n", err)
		return PrepareAuthorizationResponse{}, err
	}

	response, err := client.sendRequestWithInterface(path, "POST", headers, params)
	if err != nil {
		log.Printf("[Alipay Client] ERROR: Failed to send request: %v\n", err)
		return PrepareAuthorizationResponse{}, err
	}

	var body PrepareAuthorizationResponse
	err = json.Unmarshal(response, &body)

	if err != nil {
		log.Printf("[Alipay Client] ERROR: Failed to unmarshal response: %v\n", err)
		return PrepareAuthorizationResponse{}, err
	}

	log.Printf("[Alipay Client] Prepare response - Status: %s, Code: %s\n",
		body.Result.ResultStatus, body.Result.ResultCode)

	return body, nil
}

func (client *Client) InquiryUserCardList(accessToken string) (InquiryUserCardListResponse, error) {
	const path = "/v1/users/inquiryUserCardList"
	params := map[string]string{
		"accessToken": accessToken,
	}

	headers, err := client.buildHeaders("POST", path, params)
	if err != nil {
		return InquiryUserCardListResponse{}, err
	}

	response, err := client.sendRequest(path, "POST", headers, params)
	if err != nil {
		return InquiryUserCardListResponse{}, err
	}

	var body InquiryUserCardListResponse
	err = json.Unmarshal(response, &body)
	return body, err
}

func (client *Client) Pay(request PaymentRequest) (PaymentResponse, error) {
	const path = "/v1/payments/pay"

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return PaymentResponse{}, err
	}

	var params map[string]interface{}
	err = json.Unmarshal(requestJSON, &params)
	if err != nil {
		return PaymentResponse{}, err
	}

	headers, err := client.buildHeaders("POST", path, params)
	if err != nil {
		return PaymentResponse{}, err
	}

	response, err := client.sendRequestWithInterface(path, "POST", headers, params)
	if err != nil {
		return PaymentResponse{}, err
	}

	var body PaymentResponse
	err = json.Unmarshal(response, &body)
	return body, err
}

func (client *Client) Refund(request RefundRequest) (RefundResponse, error) {
	const path = "/v1/payments/refund"

	log.Println("[Alipay Client] Initiating refund request")
	log.Printf("[Alipay Client] Refund request ID: %s", request.RefundRequestID)
	log.Printf("[Alipay Client] Payment ID: %s", request.PaymentID)
	log.Printf("[Alipay Client] Refund amount: %s %s", request.RefundAmount.Value, request.RefundAmount.Currency)

	requestJSON, err := json.Marshal(request)
	if err != nil {
		log.Printf("[Alipay Client] ERROR: Failed to marshal refund request: %v", err)
		return RefundResponse{}, err
	}

	var params map[string]interface{}
	err = json.Unmarshal(requestJSON, &params)
	if err != nil {
		log.Printf("[Alipay Client] ERROR: Failed to unmarshal to params: %v", err)
		return RefundResponse{}, err
	}

	headers, err := client.buildHeaders("POST", path, params)
	if err != nil {
		log.Printf("[Alipay Client] ERROR: Failed to build headers: %v", err)
		return RefundResponse{}, err
	}

	response, err := client.sendRequestWithInterface(path, "POST", headers, params)
	if err != nil {
		log.Printf("[Alipay Client] ERROR: Failed to send refund request: %v", err)
		return RefundResponse{}, err
	}

	var body RefundResponse
	err = json.Unmarshal(response, &body)
	if err != nil {
		log.Printf("[Alipay Client] ERROR: Failed to unmarshal refund response: %v", err)
		return RefundResponse{}, err
	}

	log.Printf("[Alipay Client] Refund response - Status: %s, Code: %s", body.Result.ResultStatus, body.Result.ResultCode)
	return body, nil
}

func (client *Client) InquiryRefund(request InquiryRefundRequest) (InquiryRefundResponse, error) {
	const path = "/v1/payments/inquiryRefund"

	log.Println("[Alipay Client] Querying refund status")
	if request.RefundID != "" {
		log.Printf("[Alipay Client] Refund ID: %s", request.RefundID)
	}
	if request.RefundRequestID != "" {
		log.Printf("[Alipay Client] Refund Request ID: %s", request.RefundRequestID)
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		log.Printf("[Alipay Client] ERROR: Failed to marshal inquiry request: %v", err)
		return InquiryRefundResponse{}, err
	}

	var params map[string]interface{}
	err = json.Unmarshal(requestJSON, &params)
	if err != nil {
		log.Printf("[Alipay Client] ERROR: Failed to unmarshal to params: %v", err)
		return InquiryRefundResponse{}, err
	}

	headers, err := client.buildHeaders("POST", path, params)
	if err != nil {
		log.Printf("[Alipay Client] ERROR: Failed to build headers: %v", err)
		return InquiryRefundResponse{}, err
	}

	response, err := client.sendRequestWithInterface(path, "POST", headers, params)
	if err != nil {
		log.Printf("[Alipay Client] ERROR: Failed to send inquiry request: %v", err)
		return InquiryRefundResponse{}, err
	}

	var body InquiryRefundResponse
	err = json.Unmarshal(response, &body)
	if err != nil {
		log.Printf("[Alipay Client] ERROR: Failed to unmarshal inquiry response: %v", err)
		return InquiryRefundResponse{}, err
	}

	log.Printf("[Alipay Client] Inquiry response - Status: %s, Refund Status: %s", body.Result.ResultStatus, body.RefundStatus)
	return body, nil
}
