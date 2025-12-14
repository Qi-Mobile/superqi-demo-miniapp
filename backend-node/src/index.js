const express = require('express');
const cors = require('cors');
const dotenv = require('dotenv');
const { initAlipayClient } = require('./alipay');
const {
    initAuthEndpoint,
    initPaymentEndpoint,
    initRefundEndpoint,
    initAgreementEndpoint,
    initNotificationEndpoint
} = require('./api');

async function main() {
    const result = dotenv.config();
    if (result.error) {
        console.error('Error loading .env file:', result.error);
        process.exit(1);
    }

    try {
        await initAlipayClient();
    } catch (err) {
        console.error(err);
        process.exit(1);
    }

    const app = initWebServer();

    const apiRouter = express.Router();

    initAuthEndpoint(apiRouter);
    initPaymentEndpoint(apiRouter);
    initRefundEndpoint(apiRouter);
    initAgreementEndpoint(apiRouter);
    initNotificationEndpoint(apiRouter);

    app.use('/api', apiRouter);

    const port = process.env.PORT || '1999';

    app.listen(port, (err) => {
        if (err) {
            console.error('Server error:', err);
        } else {
            console.log(`Server starting on port ${port}`);
        }
    });
}

function initWebServer() {
    const app = express();

    app.use(cors());

    app.use(express.json());
    app.use(express.urlencoded({ extended: true }));

    app.use((err, req, res, next) => {
        console.error('Unhandled error:', err);
        res.status(500).json({ error: 'Internal server error' });
    });

    return app;
}

main();