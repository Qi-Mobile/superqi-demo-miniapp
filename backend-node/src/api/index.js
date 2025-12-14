const { initAuthEndpoint } = require('./authEndpoint');
const { initPaymentEndpoint } = require('./paymentEndpoint');
const { initRefundEndpoint } = require('./refundEndpoint');
const { initAgreementEndpoint } = require('./agreementEndpoint');
const { initNotificationEndpoint } = require('./notificationEndpoint');

module.exports = {
    initAuthEndpoint,
    initPaymentEndpoint,
    initRefundEndpoint,
    initAgreementEndpoint,
    initNotificationEndpoint
};