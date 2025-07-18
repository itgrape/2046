package auth

import (
	"fmt"
	"log"

	"slurm-dashboard/config"

	"github.com/go-ldap/ldap/v3"
)

func AuthenticateLDAP(cfg *config.Config, username, password string) (bool, error) {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", cfg.LDAPServerHost, cfg.LDAPServerPort))
	if err != nil {
		return false, fmt.Errorf("failed to connect to LDAP server: %w", err)
	}
	defer l.Close()

	err = l.Bind(cfg.LDAPAdminDN, cfg.LDAPAdminPassword)
	if err != nil {
		return false, fmt.Errorf("failed to bind as admin/service account: %w", err)
	}

	searchRequest := ldap.NewSearchRequest(
		cfg.LDAPSearchBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(cfg.LDAPUserSearchFilter, username),
		[]string{"dn"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return false, fmt.Errorf("user search failed: %w", err)
	}

	if len(sr.Entries) != 1 {
		log.Printf("User %s not found or not unique, entries found: %d", username, len(sr.Entries))
		return false, nil
	}

	userDN := sr.Entries[0].DN
	err = l.Bind(userDN, password)
	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return false, nil
		}
		return false, fmt.Errorf("final user bind failed: %w", err)
	}

	return true, nil
}
