package nullroute

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/sys/unix"
)

func Test_rtentry(t *testing.T) {
	tests := []struct {
		name string
		ip   net.IP
		want *RTEntry
	}{
		{
			name: "localhost",
			ip:   net.IPv4(127, 0, 0, 1),
			want: &RTEntry{
				Dst: unix.RawSockaddrInet4{
					Family: unix.AF_INET,
					Addr:   [...]byte{127, 0, 0, 1},
				},
				Genmask: unix.RawSockaddrInet4{
					Family: unix.AF_INET,
					Addr:   [...]byte{0xff, 0xff, 0xff, 0xff},
				},
				Gateway: unix.RawSockaddrInet4{
					Family: unix.AF_INET,
				},
				Flags: unix.RTF_UP | unix.RTF_HOST | unix.RTF_REJECT,
			},
		},
		{
			name: "random",
			ip:   net.IPv4(123, 123, 123, 125),
			want: &RTEntry{
				Dst: unix.RawSockaddrInet4{
					Family: unix.AF_INET,
					Addr:   [...]byte{0x7b, 0x7b, 0x7b, 0x7d},
				},
				Genmask: unix.RawSockaddrInet4{
					Family: unix.AF_INET,
					Addr:   [...]byte{0xff, 0xff, 0xff, 0xff},
				},
				Gateway: unix.RawSockaddrInet4{
					Family: unix.AF_INET,
				},
				Flags: unix.RTF_UP | unix.RTF_HOST | unix.RTF_REJECT,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rtentry(tt.ip)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("rtentry() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
