const jose = require('node-jose');

let sharedKey = Buffer.from('this_is_a_32_byte_example_key!32');

function init() {
    const jwtKey = process.env.JWT_KEY;
    if (!jwtKey || jwtKey.length === 0) {
        console.log(`Warning: JWT_KEY environment variable not set, using default value: ${sharedKey.toString()}`);
    } else {
        sharedKey = Buffer.from(jwtKey);
    }
}

init();

async function createJWE(claims) {
    try {
        const jwk = {
            kty: 'oct',
            k: sharedKey.toString('base64url')
        };

        const key = await jose.JWK.asKey(jwk);

        const claimsString = JSON.stringify(claims);
        const encrypted = await jose.JWE.createEncrypt(
            {
                format: 'compact',
                contentAlg: 'A256GCM',
                fields: {
                    alg: 'dir'
                }
            },
            key
        )
            .update(claimsString)
            .final();

        const base64Token = Buffer.from(encrypted).toString('base64');
        return base64Token;
    } catch (err) {
        throw new Error(`Error creating JWE: ${err.message}`);
    }
}

async function parseAndValidateJWE(base64Token) {
    try {
        const jweToken = Buffer.from(base64Token, 'base64').toString();

        const jwk = {
            kty: 'oct',
            k: sharedKey.toString('base64url')
        };

        const key = await jose.JWK.asKey(jwk);

        const keystore = jose.JWK.createKeyStore();
        await keystore.add(key);

        const decrypted = await jose.JWE.createDecrypt(keystore).decrypt(jweToken);

        const claims = JSON.parse(decrypted.payload.toString());
        return claims;
    } catch (err) {
        if (err.message.includes('base64')) {
            throw new Error('error decoding base64 token');
        } else if (err.message.includes('decrypt')) {
            throw new Error('unable to decrypt token');
        } else {
            throw new Error(`invalid token: ${err.message}`);
        }
    }
}

module.exports = {
    createJWE,
    parseAndValidateJWE
};