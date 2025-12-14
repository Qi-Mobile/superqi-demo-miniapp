const crypto = require('crypto');
const axios = require('axios');
const { loadPrivateKey, loadPublicKey } = require('./util');

let alipayClient = null;

class AlipayClient {
    constructor(config, privateKey, publicKey) {
        this.config = config;
        this.privateKey = privateKey;
        this.publicKey = publicKey;
        this.httpClient = axios.create({
            timeout: 25000, // 25 seconds
        });
    }

    buildHeaders(method, path, params) {
        const currentTimestamp = new Date().toISOString().replace('Z', '+00:00');
        const paramsJSON = JSON.stringify(params);

        const signature = this.generateSignature(method, path, currentTimestamp, paramsJSON);

        return {
            'Content-Type': 'application/json; charset=UTF-8',
            'Client-Id': this.config.clientId,
            'Request-Time': currentTimestamp,
            'Signature': `algorithm=RSA256, keyVersion=1, signature=${signature}`
        };
    }

    generateSignature(httpMethod, path, requestTime, content) {
        const signContent = `${httpMethod} ${path}\n${this.config.clientId}.${requestTime}.${content}`;

        const sign = crypto.createSign('SHA256');
        sign.update(signContent);
        sign.end();

        const signature = sign.sign(this.privateKey, 'base64');
        return signature;
    }

    async sendRequest(path, method, headers, params) {
        try {
            const response = await this.httpClient.request({
                url: this.config.gatewayUrl + path,
                method: method,
                headers: headers,
                data: params
            });

            return Buffer.from(JSON.stringify(response.data));
        } catch (err) {
            throw new Error(`Request failed: ${err.message}`);
        }
    }

    async sendRequestWithInterface(path, method, headers, params) {
        try {
            const response = await this.httpClient.request({
                url: this.config.gatewayUrl + path,
                method: method,
                headers: headers,
                data: params
            });

            return Buffer.from(JSON.stringify(response.data));
        } catch (err) {
            throw new Error(`Request failed: ${err.message}`);
        }
    }
}

function loadEnvConfig() {
    const gatewayUrl = process.env.ALIPAY_GATEWAY_URL;
    if (!gatewayUrl) {
        throw new Error('ALIPAY_GATEWAY_URL is not set');
    }

    const merchantPrivateKeyPath = process.env.ALIPAY_MERCHANT_PRIVATE_KEY_PATH;
    if (!merchantPrivateKeyPath) {
        throw new Error('ALIPAY_MERCHANT_PRIVATE_KEY_PATH is not set');
    }

    const alipayPublicKeyPath = process.env.ALIPAY_PUBLIC_KEY_PATH;
    if (!alipayPublicKeyPath) {
        throw new Error('ALIPAY_PUBLIC_KEY_PATH is not set');
    }

    const clientId = process.env.ALIPAY_CLIENT_ID;
    if (!clientId) {
        throw new Error('ALIPAY_CLIENT_ID is not set');
    }

    return {
        gatewayUrl,
        merchantPrivateKeyPath,
        alipayPublicKeyPath,
        clientId
    };
}

async function initAlipayClient() {
    const config = loadEnvConfig();
    const privateKey = loadPrivateKey(config.merchantPrivateKeyPath);
    const publicKey = loadPublicKey(config.alipayPublicKeyPath);

    alipayClient = new AlipayClient(config, privateKey, publicKey);
}

function getAlipayClient() {
    if (!alipayClient) {
        throw new Error('Alipay client not initialized. Call initAlipayClient() first.');
    }
    return alipayClient;
}

module.exports = {
    AlipayClient,
    initAlipayClient,
    getAlipayClient
};