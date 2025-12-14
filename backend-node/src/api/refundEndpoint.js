const { v4: uuidv4 } = require('uuid');
const { refund, inquiryRefund } = require('../alipay');

function initRefundEndpoint(router) {
    router.post('/payment/refund', async (req, res) => {
        try {
            const { paymentId, amount } = req.body;

            console.log('=================================================================');
            console.log('REFUND REQUEST RECEIVED');
            console.log('=================================================================');
            console.log(`[INFO] Payment ID: ${paymentId}`);
            console.log(`[INFO] Refund Amount (IQD): ${amount}`);

            if (!paymentId) {
                console.log('[ERROR] Payment ID is required');
                return res.status(400).json({
                    success: false,
                    resultStatus: 'F',
                    resultMessage: 'Payment ID is required'
                });
            }

            if (!amount || amount <= 0) {
                console.log('[ERROR] Invalid refund amount');
                return res.status(400).json({
                    success: false,
                    resultStatus: 'F',
                    resultMessage: 'Refund amount must be greater than 0'
                });
            }

            const refundResponse = await processRefund(paymentId, amount);
            const response = buildRefundResponse(refundResponse);

            console.log('[SUCCESS] Returning refund response to frontend');
            console.log('=================================================================');
            return res.json(response);

        } catch (err) {
            console.log(`[ERROR] Failed to process refund: ${err.message}`);
            return res.status(500).json({
                success: false,
                resultStatus: 'F',
                resultMessage: err.message
            });
        }
    });
}

async function processRefund(paymentId, amountIQD) {
    console.log('=================================================================');
    console.log(`PROCESSING REFUND FOR PAYMENT: ${paymentId}`);
    console.log('=================================================================');

    const refundRequestId = generateRefundRequestId();
    console.log(`[INFO] Generated Refund Request ID: ${refundRequestId}`);

    const amountInFils = Math.floor(amountIQD * 1000);
    console.log(`[INFO] Amount in fils: ${amountInFils}`);

    const refundRequest = {
        refundRequestId: refundRequestId,
        paymentId: paymentId,
        refundAmount: {
            currency: 'IQD',
            value: amountInFils.toString()
        },
        refundReason: 'Customer requested refund from mini app'
    };

    console.log('[INFO] Refund request details:');
    console.log(JSON.stringify(refundRequest, null, 2));
    console.log('');

    console.log('[INFO] Calling Alipay refund API...');
    let refundResponse = await refund(refundRequest);

    console.log('[SUCCESS] Refund API response:');
    console.log(JSON.stringify(refundResponse, null, 2));
    console.log('');

    switch (refundResponse.result.resultStatus) {
        case 'S':
            console.log('[SUCCESS] Refund successful immediately');
            console.log(`[INFO] Refund ID: ${refundResponse.refundId}`);
            console.log(`[INFO] Refund Time: ${refundResponse.refundTime}`);
            break;

        case 'U':
            console.log('[WARNING] Refund status unknown - starting polling...');
            const finalResponse = await pollRefundStatus(refundRequestId);
            if (finalResponse) {
                refundResponse = finalResponse;
            } else {
                console.log('[WARNING] Polling completed but status still unknown');
            }
            break;

        case 'F':
            console.log(`[ERROR] Refund failed: ${refundResponse.result.resultMessage}`);
            console.log(`[ERROR] Error Code: ${refundResponse.result.resultCode}`);
            break;
    }

    console.log('=================================================================');
    console.log('REFUND PROCESSING COMPLETED');
    console.log('=================================================================');

    return refundResponse;
}

async function pollRefundStatus(refundRequestId) {
    const maxAttempts = 12;
    const intervalSeconds = 5;

    console.log('=================================================================');
    console.log(`STARTING REFUND STATUS POLLING FOR: ${refundRequestId}`);
    console.log('=================================================================');
    console.log(`[INFO] Max attempts: ${maxAttempts}, Interval: ${intervalSeconds} seconds`);

    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
        console.log(`[INFO] Polling attempt ${attempt}/${maxAttempts}...`);

        try {
            const inquiryRequest = {
                refundRequestId: refundRequestId
            };

            const inquiryResponse = await inquiryRefund(inquiryRequest);

            console.log('[INFO] Inquiry response:');
            console.log(JSON.stringify(inquiryResponse, null, 2));
            console.log('');

            if (inquiryResponse.result.resultStatus === 'S') {
                switch (inquiryResponse.refundStatus) {
                    case 'SUCCESS':
                        console.log('[SUCCESS] Refund completed successfully!');
                        console.log(`[INFO] Refund ID: ${inquiryResponse.refundId}`);
                        console.log(`[INFO] Refund Time: ${inquiryResponse.refundTime}`);

                        return {
                            result: {
                                resultCode: 'SUCCESS',
                                resultStatus: 'S',
                                resultMessage: 'Success'
                            },
                            refundId: inquiryResponse.refundId,
                            refundTime: inquiryResponse.refundTime
                        };

                    case 'FAIL':
                        console.log(`[ERROR] Refund failed: ${inquiryResponse.refundFailReason}`);

                        return {
                            result: {
                                resultCode: 'REFUND_FAILED',
                                resultStatus: 'F',
                                resultMessage: inquiryResponse.refundFailReason
                            }
                        };

                    case 'PROCESSING':
                        console.log('[INFO] Refund still processing...');
                        break;

                    default:
                        console.log(`[WARNING] Unknown refund status: ${inquiryResponse.refundStatus}`);
                }

            } else if (inquiryResponse.result.resultStatus === 'F') {
                console.log(`[ERROR] Inquiry failed: ${inquiryResponse.result.resultMessage}`);

                if (inquiryResponse.result.resultCode === 'REFUND_NOT_EXIST') {
                    console.log('[ERROR] Refund does not exist in wallet system');
                    return {
                        result: {
                            resultCode: 'REFUND_NOT_EXIST',
                            resultStatus: 'F',
                            resultMessage: 'Refund not found in wallet system'
                        }
                    };
                }
            }

        } catch (err) {
            console.log(`[ERROR] Inquiry attempt ${attempt} failed: ${err.message}`);
        }

        if (attempt < maxAttempts) {
            console.log(`[INFO] Waiting ${intervalSeconds} seconds before next attempt...`);
            await new Promise(resolve => setTimeout(resolve, intervalSeconds * 1000));
        }
    }

    console.log('=================================================================');
    console.log('[WARNING] POLLING TIMEOUT - Refund status still unknown');
    console.log('[WARNING] Manual intervention may be required');
    console.log('=================================================================');

    return null;
}

function buildRefundResponse(refundResponse) {
    const response = {
        resultStatus: refundResponse.result.resultStatus,
        resultCode: refundResponse.result.resultCode,
        resultMessage: refundResponse.result.resultMessage
    };

    switch (refundResponse.result.resultStatus) {
        case 'S':
            response.status = 'SUCCESS';
            response.success = true;
            if (refundResponse.refundId) {
                response.refundId = refundResponse.refundId;
            }
            if (refundResponse.refundTime) {
                response.refundTime = refundResponse.refundTime;
            }
            break;

        case 'U':
            response.status = 'PENDING';
            response.success = false;
            response.message = 'Refund is being processed. Status is unknown.';
            break;

        case 'F':
            response.status = 'FAILED';
            response.success = false;
            response.message = refundResponse.result.resultMessage;
            break;
    }

    return response;
}

function generateRefundRequestId() {
    return `REFUND-${uuidv4()}-${Date.now()}`;
}

module.exports = { initRefundEndpoint };