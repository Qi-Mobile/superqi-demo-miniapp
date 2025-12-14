const crypto = require('crypto');
const fs = require('fs');

function loadPrivateKey(path) {
    try {
        const keyData = fs.readFileSync(path, 'utf8');

        const privateKey = crypto.createPrivateKey({
            key: keyData,
            format: 'pem'
        });

        return privateKey;
    } catch (err) {
        throw new Error(`Failed to load private key: ${err.message}`);
    }
}

function loadPublicKey(path) {
    try {
        const keyData = fs.readFileSync(path, 'utf8');

        const publicKey = crypto.createPublicKey({
            key: keyData,
            format: 'pem'
        });

        return publicKey;
    } catch (err) {
        throw new Error(`Failed to load public key: ${err.message}`);
    }
}

module.exports = {
    loadPrivateKey,
    loadPublicKey
};