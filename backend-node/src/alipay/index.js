const { initAlipayClient, getAlipayClient } = require('./client');

const {
    applyToken,
    inquiryUserInfo,
    prepareAuthorization,
    inquiryUserCardList,
    pay,
    refund,
    inquiryRefund,
    sendInbox,
    sendPush
} = require('./alipay');

const {
    ONLINE_PURCHASE,
    AGREEMENT_PAYMENT,
    ONLINE_PURCHASE_AUTH_CAPTURE
} = require('./constants');

module.exports = {
    initAlipayClient,
    getAlipayClient,

    applyToken,
    inquiryUserInfo,
    prepareAuthorization,
    inquiryUserCardList,
    pay,
    refund,
    inquiryRefund,
    sendInbox,
    sendPush,

    ONLINE_PURCHASE,
    AGREEMENT_PAYMENT,
    ONLINE_PURCHASE_AUTH_CAPTURE
};