sync_passwd:
  file.managed:
    - name: /etc/passwd
    - source: salt://auth_files/passwd
    - user: root
    - group: root
    - mode: 644

sync_shadow:
  file.managed:
    - name: /etc/shadow
    - source: salt://auth_files/shadow
    - user: root
    - group: root
    - mode: 644

sync_group:
  file.managed:
    - name: /etc/group
    - source: salt://auth_files/group
    - user: root
    - group: root
    - mode: 644

sync_gshadow:
  file.managed:
    - name: /etc/gshadow
    - source: salt://auth_files/gshadow
    - user: root
    - group: root
    - mode: 644

sync_login_defs:
  file.managed:
    - name: /etc/login.defs
    - source: salt://auth_files/login.defs
    - user: root
    - group: root
    - mode: 644