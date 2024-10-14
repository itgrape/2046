#!/bin/bash

IPA_SERVER=ipa.example.com
DOMAIN=example.com
REALM=EXAMPLE.COM
LOGFILE="/var/log/ipaclient-install.log"

echo "Starting FreeIPA client installation"
echo "Logfile: ${LOGFILE}"

ipa-client-install -U \
    --hostname=$(hostname) \
    --mkhomedir \
    --server=$IPA_SERVER \
    --domain=$DOMAIN \
    --realm=$REALM \
    --principal=admin \
    --password=admin_password \
    --force-join > ${LOGFILE} 2>&1

if [ $? -eq 0 ]; then
    echo "FreeIPA client installation completed successfully."
else
    echo "FreeIPA client installation failed. Check the logfile at ${LOGFILE} for details."
fi