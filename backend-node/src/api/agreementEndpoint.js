const { v4: uuidv4 } = require('uuid');
const { prepareAuthorization, applyToken, pay, AGREEMENT_PAYMENT } = require('../alipay');

function initAgreementEndpoint(router) {
    console.log('=================================================================');
    console.log('[Backend] Initializing Agreement Payment Endpoints');
    console.log('=================================================================');
    console.log('[Backend] Registering route: POST /api/agreement/prepare');
    console.log('[Backend] Registering route: POST /api/agreement/apply-token');
    console.log('[Backend] Registering route: POST /api/agreement/pay');
    console.log('=================================================================');

    router.post('/agreement/prepare', handlePrepareContract);

    router.post('/agreement/apply-token', handleApplyAccessToken);

    router.post('/agreement/pay', handleExecuteAgreementPayment);
}

async function handlePrepareContract(req, res) {
    console.log('\n=================================================================');
    console.log('[Backend] INCOMING REQUEST: POST /api/agreement/prepare');
    console.log('=================================================================');

    console.log('[Backend] Request Headers:');
    Object.keys(req.headers).forEach(key => {
        console.log(`[Backend]   ${key}: ${req.headers[key]}`);
    });

    console.log(`[Backend] Raw Request Body: ${JSON.stringify(req.body)}`);

    const { contractDescription } = req.body;

    if (!contractDescription) {
        console.log('[Backend] ERROR: contractDescription is required');
        console.log('=================================================================');
        return res.status(400).json({ error: 'contractDescription is required' });
    }

    console.log('[Backend] SUCCESS: Request parsed successfully');
    console.log(`[Backend] Contract description: ${contractDescription}`);
    console.log('[Backend] -----------------------------------------------------------');
    console.log('[Backend] Calling Alipay+ PrepareAuthorization API...');

    try {
        const prepareResponse = await prepareAuthorization(contractDescription);

        console.log('[Backend] SUCCESS: Alipay+ API call successful');
        console.log('[Backend] Alipay+ Response:');
        console.log(JSON.stringify(prepareResponse, null, 2));
        console.log('[Backend] -----------------------------------------------------------');

        console.log(`[Backend] Checking result status: ${prepareResponse.result.resultStatus}`);
        console.log(`[Backend] Result code: ${prepareResponse.result.resultCode}`);
        console.log(`[Backend] Result message: ${prepareResponse.result.resultMessage}`);

        if (prepareResponse.result.resultStatus !== 'S') {
            console.log('[Backend] ERROR: Contract preparation failed');
            console.log(`[Backend] Status: ${prepareResponse.result.resultStatus}, Code: ${prepareResponse.result.resultCode}, Message: ${prepareResponse.result.resultMessage}`);
            console.log('=================================================================');
            return res.status(400).json({
                success: false,
                resultStatus: prepareResponse.result.resultStatus,
                resultCode: prepareResponse.result.resultCode,
                resultMessage: prepareResponse.result.resultMessage
            });
        }

        if (!prepareResponse.authUrl) {
            console.log('[Backend] WARNING: authUrl is EMPTY in response!');
            console.log('[Backend] This should not happen if resultStatus is \'S\'');
            console.log(`[Backend] Full response: ${JSON.stringify(prepareResponse)}`);
            console.log('=================================================================');
        } else {
            console.log(`[Backend] SUCCESS: Authorization URL received: ${prepareResponse.authUrl}`);
        }

        const response = {
            success: true,
            authUrl: prepareResponse.authUrl,
            resultStatus: prepareResponse.result.resultStatus,
            resultCode: prepareResponse.result.resultCode,
            resultMessage: prepareResponse.result.resultMessage
        };

        console.log('[Backend] -----------------------------------------------------------');
        console.log('[Backend] Building response to frontend:');
        console.log(JSON.stringify(response, null, 2));
        console.log('[Backend] SUCCESS: Sending response to frontend');
        console.log('=================================================================');

        return res.json(response);

    } catch (err) {
        console.log(`[Backend] ERROR: Alipay+ API call failed: ${err.message}`);
        console.log('[Backend] This could mean:');
        console.log('[Backend]   1. Alipay+ gateway is unreachable');
        console.log('[Backend]   2. Invalid credentials in .env file');
        console.log('[Backend]   3. Network/firewall issue');
        console.log('[Backend]   4. PrepareAuthorization function not implemented');
        console.log('=================================================================');
        return res.status(500).json({ error: 'Failed to prepare contract: ' + err.message });
    }
}

async function handleApplyAccessToken(req, res) {
    console.log('\n=================================================================');
    console.log('[Backend] INCOMING REQUEST: POST /api/agreement/apply-token');
    console.log('=================================================================');

    console.log(`[Backend] Raw Request Body: ${JSON.stringify(req.body)}`);

    const { authCode } = req.body;

    if (!authCode) {
        console.log('[Backend] ERROR: authCode is required');
        console.log('=================================================================');
        return res.status(400).json({ error: 'authCode is required' });
    }

    console.log('[Backend] SUCCESS: Request parsed successfully');
    console.log(`[Backend] Auth code received (first 20 chars): ${truncateString(authCode, 20)}...`);
    console.log(`[Backend] Auth code length: ${authCode.length}`);
    console.log('[Backend] -----------------------------------------------------------');
    console.log('[Backend] Calling Alipay+ ApplyToken API...');

    try {
        const tokenResponse = await applyToken(authCode);

        console.log('[Backend] SUCCESS: Alipay+ API call successful');
        console.log('[Backend] Alipay+ Response:');
        console.log(JSON.stringify(tokenResponse, null, 2));
        console.log('[Backend] -----------------------------------------------------------');

        console.log(`[Backend] Checking result status: ${tokenResponse.result.resultStatus}`);
        console.log(`[Backend] Result code: ${tokenResponse.result.resultCode}`);

        if (tokenResponse.result.resultStatus !== 'S' || tokenResponse.result.resultCode !== 'SUCCESS') {
            console.log(`[Backend] ERROR: Token application failed: ${tokenResponse.result.resultMessage}`);
            console.log('=================================================================');
            return res.status(400).json({
                success: false,
                resultStatus: tokenResponse.result.resultStatus,
                resultCode: tokenResponse.result.resultCode,
                resultMessage: tokenResponse.result.resultMessage
            });
        }

        console.log(`[Backend] Access Token (first 30 chars): ${truncateString(tokenResponse.accessToken, 30)}...`);
        console.log(`[Backend] Customer ID: ${tokenResponse.customerId}`);
        console.log(`[Backend] Access Token expiry: ${tokenResponse.accessTokenExpiryTime}`);

        const response = {
            success: true,
            accessToken: tokenResponse.accessToken,
            customerId: tokenResponse.customerId,
            accessTokenExpiryTime: tokenResponse.accessTokenExpiryTime,
            resultStatus: tokenResponse.result.resultStatus,
            resultCode: tokenResponse.result.resultCode,
            resultMessage: tokenResponse.result.resultMessage
        };

        console.log('[Backend] SUCCESS: Sending token response to frontend');
        console.log('=================================================================');

        return res.json(response);

    } catch (err) {
        console.log(`[Backend] ERROR: Token exchange failed: ${err.message}`);
        console.log('=================================================================');
        return res.status(400).json({ error: 'Token exchange failed: ' + err.message });
    }
}

async function handleExecuteAgreementPayment(req, res) {
    console.log('\n=================================================================');
    console.log('[Backend] INCOMING REQUEST: POST /api/agreement/pay');
    console.log('=================================================================');

    console.log(`[Backend] Raw Request Body: ${JSON.stringify(req.body)}`);

    let { accessToken, customerId, amount, currency, orderDescription } = req.body;

    if (!currency) {
        currency = 'IQD';
    }
    if (!orderDescription) {
        orderDescription = 'Agreement payment - Monthly subscription';
    }

    console.log('[Backend] SUCCESS: Request parsed successfully');
    console.log(`[Backend] Access Token (first 20 chars): ${truncateString(accessToken, 20)}...`);
    console.log(`[Backend] Customer ID: ${customerId}`);
    console.log(`[Backend] Amount: ${amount} ${currency}`);
    console.log(`[Backend] Order Description: ${orderDescription}`);
    console.log('[Backend] -----------------------------------------------------------');

    if (!amount || amount <= 0) {
        console.log('[Backend] ERROR: Invalid payment amount');
        console.log('=================================================================');
        return res.status(400).json({
            success: false,
            resultStatus: 'F',
            resultMessage: 'Payment amount must be greater than 0'
        });
    }

    if (!accessToken) {
        console.log('[Backend] ERROR: Access token is required');
        console.log('=================================================================');
        return res.status(400).json({
            success: false,
            resultStatus: 'F',
            resultMessage: 'Access token is required'
        });
    }

    if (!customerId) {
        console.log('[Backend] ERROR: Customer ID is required');
        console.log('=================================================================');
        return res.status(400).json({
            success: false,
            resultStatus: 'F',
            resultMessage: 'Customer ID is required'
        });
    }

    console.log('[Backend] Executing agreement payment...');

    try {
        const paymentResponse = await executeAgreementPaymentInternal(
            accessToken,
            customerId,
            amount,
            currency,
            orderDescription
        );

        const response = buildAgreementPaymentResponse(paymentResponse);

        console.log('[Backend] SUCCESS: Sending payment response to frontend');
        console.log('=================================================================');
        return res.json(response);

    } catch (err) {
        console.log(`[Backend] ERROR: Failed to execute payment: ${err.message}`);
        console.log('=================================================================');
        return res.status(500).json({
            success: false,
            resultStatus: 'F',
            resultMessage: err.message
        });
    }
}

async function executeAgreementPaymentInternal(accessToken, customerId, amount, currency, orderDescription) {
    console.log('[Backend] Preparing agreement payment request...');
    console.log(`[Backend] Using Customer ID: ${customerId}`);

    const paymentRequestId = `AGREEMENT-PAY-${uuidv4()}-${Date.now()}`;
    console.log(`[Backend] Generated Payment Request ID: ${paymentRequestId}`);

    const expiryTime = new Date(Date.now() + 30 * 60 * 1000)
        .toISOString()
        .replace('Z', '+00:00');

    const baseUrl = process.env.BASE_URL || 'http://localhost:1999';

    const paymentRequest = {
        productCode: AGREEMENT_PAYMENT,
        paymentRequestId: paymentRequestId,
        paymentAuthCode: accessToken,
        paymentAmount: {
            currency: currency,
            value: amount.toString()
        },
        order: {
            orderDescription: orderDescription,
            buyer: {
                referenceBuyerId: customerId
            }
        },
        paymentExpiryTime: expiryTime,
        paymentNotifyUrl: baseUrl + '/api/webhook/payment-notify'
    };

    console.log('[Backend] Agreement payment request:');
    console.log(JSON.stringify(paymentRequest, null, 2));

    console.log('[Backend] Calling /v1/payments/pay API...');
    const paymentResponse = await pay(paymentRequest);

    console.log('[Backend] Payment API response:');
    console.log(JSON.stringify(paymentResponse, null, 2));

    switch (paymentResponse.result.resultStatus) {
        case 'S':
            console.log('[Backend] SUCCESS: Payment completed immediately');
            console.log(`[Backend] Payment ID: ${paymentResponse.paymentId}`);
            console.log(`[Backend] Payment Time: ${paymentResponse.paymentTime}`);
            console.log('[Backend] Money deducted from user\'s wallet automatically!');
            break;

        case 'U':
            console.log('[Backend] WARNING: Payment status unknown');
            console.log('[Backend] Backend should poll /v1/payments/inquiryPayment for status');
            break;

        case 'F':
            console.log(`[Backend] ERROR: Payment failed - ${paymentResponse.result.resultMessage}`);
            console.log(`[Backend] Error Code: ${paymentResponse.result.resultCode}`);
            break;

        default:
            console.log(`[Backend] WARNING: Unexpected status: ${paymentResponse.result.resultStatus}`);
    }

    return paymentResponse;
}

function buildAgreementPaymentResponse(paymentResponse) {
    const response = {
        resultStatus: paymentResponse.result.resultStatus,
        resultCode: paymentResponse.result.resultCode,
        resultMessage: paymentResponse.result.resultMessage
    };

    switch (paymentResponse.result.resultStatus) {
        case 'S':
            response.status = 'SUCCESS';
            response.success = true;
            if (paymentResponse.paymentId) {
                response.paymentId = paymentResponse.paymentId;
            }
            if (paymentResponse.paymentRequestId) {
                response.paymentRequestId = paymentResponse.paymentRequestId;
            }
            if (paymentResponse.paymentTime) {
                response.paymentTime = paymentResponse.paymentTime;
            }
            break;

        case 'U':
            response.status = 'PENDING';
            response.success = false;
            response.message = 'Payment status is unknown. Backend should poll for status.';
            if (paymentResponse.paymentId) {
                response.paymentId = paymentResponse.paymentId;
            }
            break;

        case 'F':
            response.status = 'FAILED';
            response.success = false;
            response.message = paymentResponse.result.resultMessage;
            break;

        default:
            response.status = 'UNKNOWN';
            response.success = false;
            response.message = 'Unexpected payment status: ' + paymentResponse.result.resultStatus;
    }

    return response;
}

function truncateString(s, maxLen) {
    if (!s) return '';
    if (s.length <= maxLen) {
        return s;
    }
    return s.substring(0, maxLen);
}

module.exports = { initAgreementEndpoint };