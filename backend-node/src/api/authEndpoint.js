const { applyToken, inquiryUserInfo } = require('../alipay');
const { createJWE } = require('../jwe/jwe');

function initAuthEndpoint(router) {
    router.post('/auth/apply-token', async (req, res) => {
        try {
            const { auth_code } = req.body;

            if (!auth_code) {
                return res.status(400).json({ error: 'auth_code is required' });
            }

            console.log('=================================================================');
            console.log('STARTING AUTH TOKEN EXCHANGE');
            console.log('=================================================================');

            const tokenResponse = await applyToken(auth_code);

            console.log('[SUCCESS] Token response received:');
            console.log(JSON.stringify(tokenResponse, null, 2));
            console.log('');

            if (tokenResponse.result.resultCode !== 'SUCCESS') {
                console.log(`[ERROR] Invalid token response: ${tokenResponse.result.resultMessage}`);
                return res.status(400).json({
                    error: 'Invalid token response: ' + tokenResponse.result.resultMessage
                });
            }

            const info = await inquiryUserInfo(tokenResponse.accessToken);

            console.log('[SUCCESS] User info retrieved:');
            console.log(JSON.stringify(info, null, 2));
            console.log('');

            const jweToken = await createJWE({
                user_id: info.userInfo.userId,
                access_token: tokenResponse.accessToken
            });

            console.log('[SUCCESS] Returning auth token to frontend');

            return res.json({
                token: jweToken
            });

        } catch (err) {
            console.log(`[ERROR] Auth endpoint error: ${err.message}`);
            return res.status(500).json({ error: err.message });
        }
    });
}

module.exports = { initAuthEndpoint };