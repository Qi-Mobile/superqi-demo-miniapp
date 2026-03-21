const { applyToken, inquiryUserCardList } = require('../alipay');

function initInquiryEndpoint(router) {
    // Endpoint to exchange auth code for access token (specifically for card inquiry)
    router.post('/users/inquiry-cards/apply-token', async (req, res) => {
        try {
            const { auth_code } = req.body;

            if (!auth_code) {
                return res.status(400).json({ error: 'auth_code is required' });
            }

            console.log('=================================================================');
            console.log('STARTING TOKEN EXCHANGE FOR CARD INQUIRY');
            console.log('=================================================================');
            console.log(`[INFO] Auth code received from frontend`);

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

            console.log('[INFO] Token exchange successful');
            console.log(`[INFO] Customer ID: ${tokenResponse.customerId}`);
            console.log(`[INFO] Access token obtained (valid until: ${tokenResponse.accessTokenExpiryTime})`);
            console.log('[SUCCESS] Returning access token to frontend');
            console.log('=================================================================\n');

            return res.json({
                accessToken: tokenResponse.accessToken,
                customerId: tokenResponse.customerId,
                accessTokenExpiryTime: tokenResponse.accessTokenExpiryTime
            });

        } catch (err) {
            console.log(`[ERROR] Token exchange endpoint error: ${err.message}`);
            console.log('=================================================================\n');
            return res.status(500).json({ error: err.message });
        }
    });

    // Endpoint to get user card list using access token
    router.post('/users/inquiry-cards', async (req, res) => {
        try {
            const { accessToken } = req.body;

            if (!accessToken) {
                return res.status(400).json({ error: 'accessToken is required' });
            }

            console.log('=================================================================');
            console.log('STARTING USER CARD LIST INQUIRY');
            console.log('=================================================================');
            console.log('[INFO] Access token received from frontend');
            console.log('[INFO] Calling Alipay+ inquiryUserCardList API...');

            const cardListResponse = await inquiryUserCardList(accessToken);

            console.log('[SUCCESS] Card list response received:');
            console.log(JSON.stringify(cardListResponse, null, 2));
            console.log('');

            if (cardListResponse.result.resultStatus === 'S') {
                const cardCount = cardListResponse.cardList ? cardListResponse.cardList.length : 0;
                console.log(`[SUCCESS] Card inquiry successful - ${cardCount} card(s) found`);

                if (cardCount > 0) {
                    console.log('[INFO] Card details:');
                    cardListResponse.cardList.forEach((card, index) => {
                        console.log(`  Card ${index + 1}:`);
                        console.log(`    Masked Card No: ${card.maskedCardNo}`);
                        console.log(`    Account Number: ${card.accountNumber}`);
                    });
                } else {
                    console.log('[INFO] User has no cards bound to their account');
                }
            } else if (cardListResponse.result.resultStatus === 'F') {
                console.log(`[ERROR] Card inquiry failed: ${cardListResponse.result.resultMessage}`);
                console.log(`[ERROR] Result code: ${cardListResponse.result.resultCode}`);
            } else if (cardListResponse.result.resultStatus === 'U') {
                console.log(`[WARNING] Card inquiry status unknown: ${cardListResponse.result.resultMessage}`);
            }

            console.log('[SUCCESS] Returning card list to frontend');
            console.log('=================================================================\n');

            return res.json(cardListResponse);

        } catch (err) {
            console.log(`[ERROR] Card inquiry endpoint error: ${err.message}`);
            console.log('=================================================================\n');
            return res.status(500).json({ error: err.message });
        }
    });
}

module.exports = { initInquiryEndpoint };
