package main

import (
	"fmt"
	"net/netip"
	"regexp"
	"strings"
)

var domainLabel = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

func validateAllowedDomains(value string) ([]string, error) {
	allowedDomains := strings.Split(value, ",")
	for index, domain := range allowedDomains {
		domain = strings.TrimSpace(domain)
		if !isNetworkDomain(domain) {
			return nil, fmt.Errorf("%s must list comma-separated domains: invalid domain %q", allowedDomainsEnv, domain)
		}
		allowedDomains[index] = domain
	}
	return allowedDomains, nil
}

func isNetworkDomain(domain string) bool {
	if len(domain) > 253 {
		return false
	}
	if _, err := netip.ParseAddr(domain); err == nil {
		return false
	}
	for label := range strings.SplitSeq(domain, ".") {
		if !domainLabel.MatchString(label) {
			return false
		}
	}
	return true
}
