const { v4: uuidv4 } = require('uuid');
const { generateShortLink } = require('../alipay');

const DEFAULT_DESCRIPTION = 'Miniapp short link';

function initShortLinkEndpoint(router) {
    // Endpoint to generate a short link (QR code link) that opens a miniapp page
    router.post('/miniapps/short-link', async (req, res) => {
        try {
            const { appId, pagePath, queryParams, description } = req.body;

            if (!appId) {
                return res.status(400).json({ error: 'appId is required' });
            }

            // pagePath and queryParams are optional: without a page the link opens the miniapp's default page
            if (pagePath && !pagePath.startsWith('/')) {
                return res.status(400).json({ error: 'pagePath must start with /' });
            }

            console.log('=================================================================');
            console.log('STARTING SHORT LINK GENERATION');
            console.log('=================================================================');

            const requestId = `SHORTLINK-${uuidv4()}-${Date.now()}`;
            console.log(`[INFO] Request ID: ${requestId}`);
            console.log(`[INFO] App ID: ${appId}, page: ${pagePath || '(default)'}, params: ${queryParams || ''}`);

            // The gateway rejects requests without a description
            const response = await generateShortLink(requestId, appId, pagePath || '', queryParams || '', description || DEFAULT_DESCRIPTION);

            console.log('[INFO] Short link response received:');
            console.log(JSON.stringify(response, null, 2));

            if (response.result && response.result.resultStatus === 'S') {
                console.log(`[SUCCESS] Short link generated: ${response.appQrCode}`);
            } else {
                console.log(`[ERROR] Short link generation failed: ${response.result?.resultMessage} (${response.result?.resultCode})`);
            }
            console.log('=================================================================\n');

            return res.json(response);

        } catch (err) {
            console.log(`[ERROR] Short link endpoint error: ${err.message}`);
            console.log('=================================================================\n');
            return res.status(500).json({ error: err.message });
        }
    });
}

module.exports = { initShortLinkEndpoint };
