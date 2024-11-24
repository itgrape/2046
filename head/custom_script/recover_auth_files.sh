\cp -rf /srv/salt/recover/* /etc/
salt '*' state.apply sync_auth_files