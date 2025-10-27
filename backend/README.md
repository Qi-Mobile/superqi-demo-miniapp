# Sample Mini App - Backend

This directory contains the backend components of the Sample Mini App, including API endpoints and authentication services

## Overview

A Go client for integrating with SuperQi's API services. Provides secure authentication, request signing, and methods for token management, user information retrieval, and card list inquiries.

## Configuration

Requires the following environment variables:
- `SUPERQI_GATEWAY_URL`: SuperQi API gateway URL
- `SUPERQI_MERCHANT_PRIVATE_KEY_PATH`: Path to merchant's RSA private key
- `SUPERQI_PUBLIC_KEY_PATH`: Path to merchant's RSA public key  
- `SUPERQI_CLIENT_ID`: Your SuperQi client ID

## API Methods

### ApplyToken
Exchanges authorization code for access and refresh tokens.

### InquiryUserInfo  
Retrieves user profile information using access token. Returns user details like name, contact info, and preferences based on granted scopes.

### InquiryUserCardList
Gets user's linked payment cards. Requires `CARD_LIST` scope in authorization.
