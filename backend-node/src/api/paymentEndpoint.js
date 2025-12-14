const { v4: uuidv4 } = require('uuid');
const { pay } = require('../alipay');
const { ONLINE_PURCHASE } = require('../alipay');
const { parseAndValidateJWE } = require('../jwe/jwe');

function initPaymentEndpoint(router) {
    router.post('/payment/create', async (req, res) => {
        try {
            const { token } = req.body;

            if (!token) {
                return res.status(400).json({ error: 'token is required' });
            }

            console.log('=================================================================');
            console.log('PAYMENT CREATION REQUEST RECEIVED');
            console.log('=================================================================');

            const claims = await parseAndValidateJWE(token);

            console.log(`[INFO] Creating payment for user ID: ${claims.user_id}`);

            const paymentResponse = await createTestPayment(claims.user_id);

            const response = {
                success: true,
                amount: 1
            };

            if (paymentResponse.redirectActionForm && paymentResponse.redirectActionForm.redirectUrl) {
                response.paymentUrl = paymentResponse.redirectActionForm.redirectUrl;
                response.paymentId = paymentResponse.paymentId;
                console.log(`[INFO] Sending payment URL to frontend: ${paymentResponse.redirectActionForm.redirectUrl}`);
            } else {
                console.log('[WARNING] No payment URL in response');
                response.success = false;
                response.error = 'No redirect URL received from payment API';
            }

            console.log('[SUCCESS] Returning payment response to frontend');
            return res.json(response);

        } catch (err) {
            console.log(`[ERROR] Payment endpoint error: ${err.message}`);
            return res.status(500).json({ error: err.message });
        }
    });
}

async function createTestPayment(userId) {
    console.log('=================================================================');
    console.log(`CREATING TEST PAYMENT FOR USER: ${userId}`);
    console.log('=================================================================');

    const paymentRequestId = `PAY-${uuidv4()}-${Date.now()}`;

    const expiryTime = new Date(Date.now() + 30 * 60 * 1000).toISOString().replace(/\.\d{3}Z$/, '+00:00');

    const baseUrl = process.env.BASE_URL || 'http://localhost:1999';

    const paymentRequest = {
        productCode: ONLINE_PURCHASE,
        paymentRequestId: paymentRequestId,
        paymentAmount: {
            currency: 'IQD',
            value: '1000'
        },
        order: {
            orderDescription: 'Test Order - Online Purchase',
            buyer: {
                referenceBuyerId: userId
            }
        },
        paymentExpiryTime: expiryTime,
        paymentRedirectUrl: baseUrl + '/payment-success.html'
    };

    console.log('[INFO] Payment request details:');
    console.log(JSON.stringify(paymentRequest, null, 2));
    console.log('');

    console.log('[INFO] Calling payment API...');
    const paymentResponse = await pay(paymentRequest);

    console.log('[SUCCESS] Payment API response received:');
    console.log(JSON.stringify(paymentResponse, null, 2));
    console.log('');

    const redirectUrl = paymentResponse.redirectActionForm?.redirectUrl || '';

    if (paymentResponse.result.resultStatus === 'A') {
        console.log('[SUCCESS] Payment accepted');
        if (redirectUrl) {
            console.log(`[INFO] Redirection URL: ${redirectUrl}`);
            console.log('[INFO] Frontend should call my.tradePay() with this URL');
        } else {
            console.log('[WARNING] Redirection URL is empty in response');
        }
        console.log(`[INFO] Payment ID: ${paymentResponse.paymentId}`);
        console.log(`[INFO] Payment Request ID: ${paymentResponse.paymentRequestId}`);
    } else if (paymentResponse.result.resultStatus === 'S') {
        console.log('[SUCCESS] Payment completed immediately');
    } else if (paymentResponse.result.resultStatus === 'U') {
        console.log('[WARNING] Unknown payment status - need to query later');
    } else {
        console.log(`[ERROR] Payment failed: ${paymentResponse.result.resultMessage}`);
    }

    console.log('=================================================================');
    console.log('TEST PAYMENT CREATION COMPLETED');
    console.log('=================================================================');

    return paymentResponse;
}

module.exports = { initPaymentEndpoint };
