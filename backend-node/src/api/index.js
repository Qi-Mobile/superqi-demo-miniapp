const { initAuthEndpoint } = require('./authEndpoint');
const { initPaymentEndpoint } = require('./paymentEndpoint');
const { initRefundEndpoint } = require('./refundEndpoint');
const { initAgreementEndpoint } = require('./agreementEndpoint');
const { initNotificationEndpoint } = require('./notificationEndpoint');
const { initInquiryEndpoint } = require('./inquiryEndpoint');

module.exports = {
    initAuthEndpoint,
    initPaymentEndpoint,
    initRefundEndpoint,
    initAgreementEndpoint,
    initNotificationEndpoint,
    initInquiryEndpoint
};