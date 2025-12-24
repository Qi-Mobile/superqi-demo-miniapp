const { getAlipayClient } = require('./client');

async function applyToken(authCode) {
    const client = getAlipayClient();
    const path = '/v1/authorizations/applyToken';
    const params = {
        grantType: 'AUTHORIZATION_CODE',
        authCode: authCode
    };

    const headers = client.buildHeaders('POST', path, params);
    const response = await client.sendRequest(path, 'POST', headers, params);

    return JSON.parse(response.toString());
}

async function inquiryUserInfo(accessToken) {
    const client = getAlipayClient();
    const path = '/v1/users/inquiryUserInfo';
    const params = {
        accessToken: accessToken
    };

    const headers = client.buildHeaders('POST', path, params);
    const response = await client.sendRequest(path, 'POST', headers, params);

    return JSON.parse(response.toString());
}

async function prepareAuthorization(contractDescription) {
    const client = getAlipayClient();
    const path = '/v1/authorizations/prepare';

    const extendInfoMap = {
        language: 'en-US',
        contractDesc: contractDescription
    };

    const params = {
        scopes: 'AGREEMENT_PAY',
        extendInfo: JSON.stringify(extendInfoMap)
    };

    console.log('[Alipay Client] Preparing authorization request');
    console.log(`[Alipay Client] Scopes: ${params.scopes}`);
    console.log(`[Alipay Client] ExtendInfo: ${params.extendInfo}`);

    try {
        const headers = client.buildHeaders('POST', path, params);
        const response = await client.sendRequestWithInterface(path, 'POST', headers, params);
        const body = JSON.parse(response.toString());

        console.log(`[Alipay Client] Prepare response - Status: ${body.result.resultStatus}, Code: ${body.result.resultCode}`);
        return body;
    } catch (err) {
        console.log(`[Alipay Client] ERROR: ${err.message}`);
        throw err;
    }
}

async function inquiryUserCardList(accessToken) {
    const client = getAlipayClient();
    const path = '/v1/users/inquiryUserCardList';
    const params = {
        accessToken: accessToken
    };

    const headers = client.buildHeaders('POST', path, params);
    const response = await client.sendRequest(path, 'POST', headers, params);

    return JSON.parse(response.toString());
}

async function pay(request) {
    const client = getAlipayClient();
    const path = '/v1/payments/pay';

    const headers = client.buildHeaders('POST', path, request);
    const response = await client.sendRequestWithInterface(path, 'POST', headers, request);

    return JSON.parse(response.toString());
}

async function refund(request) {
    const client = getAlipayClient();
    const path = '/v1/payments/refund';

    console.log('[Alipay Client] Initiating refund request');
    console.log(`[Alipay Client] Refund request ID: ${request.refundRequestId}`);
    console.log(`[Alipay Client] Payment ID: ${request.paymentId}`);
    console.log(`[Alipay Client] Refund amount: ${request.refundAmount.value} ${request.refundAmount.currency}`);

    try {
        const headers = client.buildHeaders('POST', path, request);
        const response = await client.sendRequestWithInterface(path, 'POST', headers, request);
        const body = JSON.parse(response.toString());

        console.log(`[Alipay Client] Refund response - Status: ${body.result.resultStatus}, Code: ${body.result.resultCode}`);
        return body;
    } catch (err) {
        console.log(`[Alipay Client] ERROR: ${err.message}`);
        throw err;
    }
}

async function inquiryRefund(request) {
    const client = getAlipayClient();
    const path = '/v1/payments/inquiryRefund';

    console.log('[Alipay Client] Querying refund status');
    if (request.refundId) {
        console.log(`[Alipay Client] Refund ID: ${request.refundId}`);
    }
    if (request.refundRequestId) {
        console.log(`[Alipay Client] Refund Request ID: ${request.refundRequestId}`);
    }

    try {
        const headers = client.buildHeaders('POST', path, request);
        const response = await client.sendRequestWithInterface(path, 'POST', headers, request);
        const body = JSON.parse(response.toString());

        console.log(`[Alipay Client] Inquiry response - Status: ${body.result.resultStatus}, Refund Status: ${body.refundStatus}`);
        return body;
    } catch (err) {
        console.log(`[Alipay Client] ERROR: ${err.message}`);
        throw err;
    }
}

async function sendInbox(request) {
    const client = getAlipayClient();
    const path = '/v1/messages/sendInbox';

    console.log('[Alipay Client] Sending inbox notification');
    console.log(`[Alipay Client] Request ID: ${request.requestId}`);
    console.log(`[Alipay Client] Template Code: ${request.templateCode}`);

    try {
        const headers = client.buildHeaders('POST', path, request);
        const response = await client.sendRequestWithInterface(path, 'POST', headers, request);
        const body = JSON.parse(response.toString());

        console.log(`[Alipay Client] SendInbox response - Status: ${body.result.resultStatus}, Code: ${body.result.resultCode}`);
        return body;
    } catch (err) {
        console.log(`[Alipay Client] ERROR: ${err.message}`);
        throw err;
    }
}

async function sendPush(request) {
    const client = getAlipayClient();
    const path = '/v1/messages/sendPush';

    console.log('[Alipay Client] Sending push notification');
    console.log(`[Alipay Client] Request ID: ${request.requestId}`);
    console.log(`[Alipay Client] Template Code: ${request.templateCode}`);

    try {
        const headers = client.buildHeaders('POST', path, request);
        const response = await client.sendRequestWithInterface(path, 'POST', headers, request);
        const body = JSON.parse(response.toString());

        console.log(`[Alipay Client] SendPush response - Status: ${body.result.resultStatus}, Code: ${body.result.resultCode}`);
        return body;
    } catch (err) {
        console.log(`[Alipay Client] ERROR: ${err.message}`);
        throw err;
    }
}

module.exports = {
    applyToken,
    inquiryUserInfo,
    prepareAuthorization,
    inquiryUserCardList,
    pay,
    refund,
    inquiryRefund,
    sendInbox,
    sendPush
};