package vault

import "testing"

func TestIsAllowed_EmptyLists(t *testing.T) {
	if !IsAllowed("10.0.0.1", nil, nil) {
		t.Fatal("empty allow/deny lists must allow all traffic")
	}
	if !IsAllowed("203.0.113.5", []string{}, []string{}) {
		t.Fatal("empty allow/deny lists must allow all traffic")
	}
}

func TestIsAllowed_AllowListMatch(t *testing.T) {
	if !IsAllowed("10.1.2.3", []string{"10.0.0.0/8"}, nil) {
		t.Fatal("10.1.2.3 should be in 10.0.0.0/8")
	}
	if IsAllowed("192.168.1.1", []string{"10.0.0.0/8"}, nil) {
		t.Fatal("192.168.1.1 should NOT be in 10.0.0.0/8")
	}
}

func TestIsAllowed_DenyOverridesAllow(t *testing.T) {
	if IsAllowed("10.1.2.3", []string{"10.0.0.0/8"}, []string{"10.1.2.0/24"}) {
		t.Fatal("denied subnet must override allowed supernet")
	}
}

func TestIsAllowed_BareIPInAllowList(t *testing.T) {
	if !IsAllowed("192.168.1.5", []string{"192.168.1.5"}, nil) {
		t.Fatal("bare IP in allow list should match exactly")
	}
}

func TestIsAllowed_InvalidCIDRNotMatched(t *testing.T) {
	// An allow list with only an invalid entry means "no entry matched",
	// which we treat as deny. This is a safer default than silently
	// treating invalid input as no-restriction.
	if IsAllowed("10.0.0.1", []string{"not-a-cidr"}, nil) {
		t.Fatal("an allow list with only invalid entries must deny traffic")
	}
}

func TestIsAllowed_IPv6CIDR(t *testing.T) {
	if !IsAllowed("2001:db8::1", []string{"2001:db8::/32"}, nil) {
		t.Fatal("IPv6 address should match IPv6 CIDR")
	}
	if IsAllowed("::1", []string{"2001:db8::/32"}, nil) {
		t.Fatal("::1 should NOT be in 2001:db8::/32")
	}
}
