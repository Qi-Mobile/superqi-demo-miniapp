package alipay

import "time"

type Result struct {
	ResultCode    string `json:"resultCode"`
	ResultStatus  string `json:"resultStatus"`
	ResultMessage string `json:"resultMessage"`
}

type ApplyTokenResponse struct {
	Result                 Result    `json:"result"`
	AccessToken            string    `json:"accessToken"`
	AccessTokenExpiryTime  time.Time `json:"accessTokenExpiryTime"`
	RefreshToken           string    `json:"refreshToken"`
	RefreshTokenExpiryTime time.Time `json:"refreshTokenExpiryTime"`
	CustomerID             string    `json:"customerId"`
}

type InquiryUserCardListResponse struct {
	Result   Result `json:"result"`
	CardList []struct {
		MaskedCardNo  string `json:"maskedCardNo"`
		AccountNumber string `json:"accountNumber"`
	} `json:"cardList"`
}

type InquiryUserInfoResponse struct {
	Result   Result `json:"result"`
	UserInfo struct {
		UserID       string `json:"userId"`
		LoginIDInfos []struct {
			LoginID     string `json:"loginId"`
			HashLoginID string `json:"hashLoginId"`
			MaskLoginID string `json:"maskLoginId"`
			LoginIDType string `json:"loginIdType"`
		} `json:"loginIdInfos"`
		UserName struct {
			FullName   string `json:"fullName"`
			FirstName  string `json:"firstName"`
			SecondName string `json:"secondName"`
			ThirdName  string `json:"thirdName"`
			LastName   string `json:"lastName"`
		} `json:"userName"`
		UserNameInArabic struct {
			FullName   string `json:"fullName"`
			FirstName  string `json:"firstName"`
			SecondName string `json:"secondName"`
			ThirdName  string `json:"thirdName"`
			LastName   string `json:"lastName"`
		} `json:"userNameInArabic"`
		Avatar       string `json:"avatar"`
		Gender       string `json:"gender"`
		BirthDate    string `json:"birthDate"`
		Nationality  string `json:"nationality"`
		ContactInfos []struct {
			ContactType string `json:"contactType"`
			ContactNo   string `json:"contactNo"`
		} `json:"contactInfos"`
	} `json:"userInfo"`
}

// Payment related types below
type PaymentAmount struct {
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

type OrderBuyer struct {
	ReferenceBuyerID string `json:"referenceBuyerId"`
}

type Order struct {
	OrderDescription string     `json:"orderDescription"`
	Buyer            OrderBuyer `json:"buyer"`
}

type PaymentRequest struct {
	ProductCode        string        `json:"productCode"`
	PaymentRequestID   string        `json:"paymentRequestId"`
	PaymentAuthCode    string        `json:"paymentAuthCode,omitempty"`
	PaymentAmount      PaymentAmount `json:"paymentAmount"`
	Order              Order         `json:"order,omitempty"`
	PaymentExpiryTime  string        `json:"paymentExpiryTime,omitempty"`
	PaymentNotifyURL   string        `json:"paymentNotifyUrl,omitempty"`
	PaymentRedirectURL string        `json:"paymentRedirectUrl,omitempty"`
}

type RedirectActionForm struct {
	RedirectURL string `json:"redirectUrl"`
	Method      string `json:"method,omitempty"`
}

type PaymentResponse struct {
	Result             Result             `json:"result"`
	PaymentID          string             `json:"paymentId"`
	PaymentRequestID   string             `json:"paymentRequestId"`
	PaymentTime        string             `json:"paymentTime,omitempty"`
	RedirectActionForm RedirectActionForm `json:"redirectActionForm,omitempty"`
}

// below method can be neglected
// Helper method to get the redirect URL
func (pr *PaymentResponse) GetRedirectURL() string {
	return pr.RedirectActionForm.RedirectURL
}

// refund related types below
type RefundAmount struct {
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

type RefundRequest struct {
	RefundRequestID  string       `json:"refundRequestId"`
	PaymentID        string       `json:"paymentId,omitempty"`
	PaymentRequestID string       `json:"paymentRequestId,omitempty"`
	CaptureID        string       `json:"captureId,omitempty"`
	RefundAmount     RefundAmount `json:"refundAmount"`
	RefundReason     string       `json:"refundReason,omitempty"`
	ExtendInfo       string       `json:"extendInfo,omitempty"`
}

type RefundResponse struct {
	Result     Result `json:"result"`
	RefundID   string `json:"refundId,omitempty"`
	RefundTime string `json:"refundTime,omitempty"`
}

type InquiryRefundRequest struct {
	RefundID        string `json:"refundId,omitempty"`
	RefundRequestID string `json:"refundRequestId,omitempty"`
}

type InquiryRefundResponse struct {
	Result           Result       `json:"result"`
	RefundID         string       `json:"refundId,omitempty"`
	RefundRequestID  string       `json:"refundRequestId,omitempty"`
	RefundAmount     RefundAmount `json:"refundAmount,omitempty"`
	RefundReason     string       `json:"refundReason,omitempty"`
	RefundTime       string       `json:"refundTime,omitempty"`
	RefundStatus     string       `json:"refundStatus,omitempty"`
	RefundFailReason string       `json:"refundFailReason,omitempty"`
}

// Agreement Payment - Prepare Authorization Response
type PrepareAuthorizationResponse struct {
	Result  Result `json:"result"`
	AuthURL string `json:"authUrl"`
}
