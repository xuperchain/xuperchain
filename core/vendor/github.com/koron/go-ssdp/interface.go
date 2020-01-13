package ssdp

import "net"

// Interfaces specify target interfaces to multicast.  If no interfaces are
// specified, all interfaces will be used.
var Interfaces []net.Interface

func interfaces() []net.Interface {
	if Interfaces == nil {
		Interfaces = interfacesIPv4()
	}
	return Interfaces
}

func interfacesIPv4() []net.Interface {
	iflist, err := net.Interfaces()
	if err != nil {
		logf("failed to list interfaces: %s", err)
		return make([]net.Interface, 0)
	}
	list := make([]net.Interface, 0, len(iflist))
	for _, ifi := range iflist {
		if !hasLinkUp(&ifi) || !hasIPv4Address(&ifi) {
			continue
		}
		list = append(list, ifi)
	}
	return list
}

// hasLinkUp checks an I/F have link-up or not.
func hasLinkUp(ifi *net.Interface) bool {
	return ifi.Flags&net.FlagUp != 0
}

// hasIPv4Address checks an I/F have IPv4 address.
func hasIPv4Address(ifi *net.Interface) bool {
	addrs, err := ifi.Addrs()
	if err != nil {
		return false
	}
	for _, a := range addrs {
		ip, _, err := net.ParseCIDR(a.String())
		if err != nil {
			continue
		}
		if len(ip.To4()) == net.IPv4len && !ip.IsUnspecified() {
			return true
		}
	}
	return false
}
