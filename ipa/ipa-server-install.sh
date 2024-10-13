#!/bin/bash

IPA_SERVER_HOSTNAME=ipa.example.com
IPA_SERVER_IP=ipa.example.com
DOMAIN=example.com
REALM=EXAMPLE.COM

ipa-server-install -U \
    --hostname=${IPA_SERVER_HOSTNAME} \
    --domain=${DOMAIN} \
    --realm=${REALM} \
    --ds-password=admin_password \
    --admin-password=admin_password \
    --ip-address=${IPA_SERVER_IP} \
    --setup-dns \
    --no-forwarders
