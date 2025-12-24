const { v4: uuidv4 } = require('uuid');
const { sendInbox, sendPush } = require('../alipay');
const { parseAndValidateJWE } = require('../jwe/jwe');

function initNotificationEndpoint(router) {
    router.post('/notification/send-inbox', async (req, res) => {
        try {
            const { token, title, content, url } = req.body;

            if (!token) {
                return res.status(400).json({ error: 'token is required' });
            }
            if (!title) {
                return res.status(400).json({ error: 'title is required' });
            }
            if (!content) {
                return res.status(400).json({ error: 'content is required' });
            }

            console.log('=================================================================');
            console.log('SEND INBOX NOTIFICATION REQUEST RECEIVED');
            console.log('=================================================================');

            const claims = await parseAndValidateJWE(token);

            console.log(`[INFO] Sending notification for user ID: ${claims.user_id}`);
            console.log(`[INFO] Title: ${title}`);
            console.log(`[INFO] Content: ${content}`);

            const notificationResponse = await sendInboxNotification(claims.access_token, title, content, url);
            const response = buildNotificationResponse(notificationResponse);

            console.log('[SUCCESS] Returning notification response to frontend');
            console.log('=================================================================');
            return res.json(response);

        } catch (err) {
            console.log(`[ERROR] Failed to send notification: ${err.message}`);
            return res.status(500).json({
                success: false,
                error: err.message
            });
        }
    });

    router.post('/notification/send-push', async (req, res) => {
        try {
            const { token, title, content, url } = req.body;

            if (!token) {
                return res.status(400).json({ error: 'token is required' });
            }
            if (!title) {
                return res.status(400).json({ error: 'title is required' });
            }
            if (!content) {
                return res.status(400).json({ error: 'content is required' });
            }

            console.log('=================================================================');
            console.log('SEND PUSH NOTIFICATION REQUEST RECEIVED');
            console.log('=================================================================');

            const claims = await parseAndValidateJWE(token);

            console.log(`[INFO] Sending push notification for user ID: ${claims.user_id}`);
            console.log(`[INFO] Title: ${title}`);
            console.log(`[INFO] Content: ${content}`);

            const pushResponse = await sendPushNotification(claims.access_token, title, content, url);
            const response = buildPushNotificationResponse(pushResponse);

            console.log('[SUCCESS] Returning push notification response to frontend');
            console.log('=================================================================');
            return res.json(response);

        } catch (err) {
            console.log(`[ERROR] Failed to send push notification: ${err.message}`);
            return res.status(500).json({
                success: false,
                error: err.message
            });
        }
    });
}

async function sendInboxNotification(accessToken, title, content, url) {
    console.log('=================================================================');
    console.log('PROCESSING INBOX NOTIFICATION');
    console.log('=================================================================');

    const requestId = generateNotificationRequestId();
    console.log(`[INFO] Generated Request ID: ${requestId}`);

    if (!url) {
        url = 'mini://platformapi/startapp?_ariver_appid=888888';
    }

    const templateParams = {
        Title: title,
        Content: content,
        Url: url
    };

    const notificationRequest = {
        accessToken: accessToken,
        requestId: requestId,
        templateCode: 'MINI_APP_COMMON_INBOX',
        templates: [
            {
                templateParameters: templateParams
            }
        ]
    };

    console.log('[INFO] Notification request details:');
    console.log(JSON.stringify(notificationRequest, null, 2));
    console.log('');

    console.log('[INFO] Calling Alipay SendInbox API...');
    const notificationResponse = await sendInbox(notificationRequest);

    console.log('[SUCCESS] SendInbox API response:');
    console.log(JSON.stringify(notificationResponse, null, 2));
    console.log('');

    switch (notificationResponse.result.resultStatus) {
        case 'S':
            console.log('[SUCCESS] Notification sent successfully');
            if (notificationResponse.messageId) {
                console.log(`[INFO] Message ID: ${notificationResponse.messageId}`);
            }
            break;

        case 'A':
            console.log('[SUCCESS] Notification accepted by wallet');
            break;

        case 'U':
            console.log('[WARNING] Notification status unknown');
            break;

        case 'F':
            console.log(`[ERROR] Notification failed: ${notificationResponse.result.resultMessage}`);
            console.log(`[ERROR] Error Code: ${notificationResponse.result.resultCode}`);
            break;
    }

    console.log('=================================================================');
    console.log('NOTIFICATION PROCESSING COMPLETED');
    console.log('=================================================================');

    return notificationResponse;
}

function buildNotificationResponse(notificationResponse) {
    const response = {
        resultStatus: notificationResponse.result.resultStatus,
        resultCode: notificationResponse.result.resultCode,
        resultMessage: notificationResponse.result.resultMessage
    };

    switch (notificationResponse.result.resultStatus) {
        case 'S':
        case 'A':
            response.status = 'SUCCESS';
            response.success = true;
            if (notificationResponse.messageId) {
                response.messageId = notificationResponse.messageId;
            }
            if (notificationResponse.extendInfo) {
                response.extendInfo = notificationResponse.extendInfo;
            }
            break;

        case 'U':
            response.status = 'UNKNOWN';
            response.success = false;
            response.message = 'Notification status is unknown. It may still be processed.';
            break;

        case 'F':
            response.status = 'FAILED';
            response.success = false;
            response.message = notificationResponse.result.resultMessage;
            break;
    }

    return response;
}

async function sendPushNotification(accessToken, title, content, url) {
    console.log('=================================================================');
    console.log('PROCESSING PUSH NOTIFICATION');
    console.log('=================================================================');

    const requestId = generateNotificationRequestId();
    console.log(`[INFO] Generated Request ID: ${requestId}`);

    if (!url) {
        url = 'mini://platformapi/startapp?_ariver_appid=888888';
    }

    const templateParams = {
        Title: title,
        Content: content,
        Url: url
    };

    const pushRequest = {
        accessToken: accessToken,
        requestId: requestId,
        templateCode: 'MINI_APP_COMMON_PUSH',
        templates: [
            {
                templateParameters: templateParams
            }
        ]
    };

    console.log('[INFO] Push notification request details:');
    console.log(JSON.stringify(pushRequest, null, 2));
    console.log('');

    console.log('[INFO] Calling Alipay SendPush API...');
    const pushResponse = await sendPush(pushRequest);

    console.log('[SUCCESS] SendPush API response:');
    console.log(JSON.stringify(pushResponse, null, 2));
    console.log('');

    switch (pushResponse.result.resultStatus) {
        case 'S':
            console.log('[SUCCESS] Push notification sent successfully');
            if (pushResponse.messageId) {
                console.log(`[INFO] Message ID: ${pushResponse.messageId}`);
            }
            break;

        case 'A':
            console.log('[SUCCESS] Push notification accepted by wallet');
            break;

        case 'U':
            console.log('[WARNING] Push notification status unknown');
            break;

        case 'F':
            console.log(`[ERROR] Push notification failed: ${pushResponse.result.resultMessage}`);
            console.log(`[ERROR] Error Code: ${pushResponse.result.resultCode}`);
            break;
    }

    console.log('=================================================================');
    console.log('PUSH NOTIFICATION PROCESSING COMPLETED');
    console.log('=================================================================');

    return pushResponse;
}

function buildPushNotificationResponse(pushResponse) {
    const response = {
        resultStatus: pushResponse.result.resultStatus,
        resultCode: pushResponse.result.resultCode,
        resultMessage: pushResponse.result.resultMessage
    };

    switch (pushResponse.result.resultStatus) {
        case 'S':
        case 'A':
            response.status = 'SUCCESS';
            response.success = true;
            if (pushResponse.messageId) {
                response.messageId = pushResponse.messageId;
            }
            if (pushResponse.extendInfo) {
                response.extendInfo = pushResponse.extendInfo;
            }
            break;

        case 'U':
            response.status = 'UNKNOWN';
            response.success = false;
            response.message = 'Push notification status is unknown. It may still be processed.';
            break;

        case 'F':
            response.status = 'FAILED';
            response.success = false;
            response.message = pushResponse.result.resultMessage;
            break;
    }

    return response;
}

function generateNotificationRequestId() {
    return `NOTIF-${uuidv4()}-${Date.now()}`;
}

module.exports = { initNotificationEndpoint };
