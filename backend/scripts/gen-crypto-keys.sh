#!/bin/bash
echo "FIELD_ENCRYPTION_KEY=$(openssl rand -base64 32)"
echo "FIELD_HMAC_KEY=$(openssl rand -base64 32)"
