#!/bin/bash

IPA_SERVER=ipa.example.com
DOMAIN=example.com
REALM=EXAMPLE.COM

ipa-client-install -U \
    --hostname=$(hostname) \
    --mkhomedir \
    --server=$IPA_SERVER \
    --domain=$DOMAIN \
    --realm=$REALM \
    --principal=admin \
    --password=admin_password \
    --force-join
