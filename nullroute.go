package nullroute

import (
	"net"
	"unsafe"

	"golang.org/x/sys/unix"
)

type NullRoute int

func New() (NullRoute, error) {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_IP)
	return NullRoute(fd), err
}

func (fd NullRoute) Close() error {
	return unix.Close(int(fd))
}

func (fd NullRoute) Add(ip net.IP) error {
	return IoctlAddRoute(int(fd), rtentry(ip))
}

func (fd NullRoute) Delete(ip net.IP) error {
	return IoctlDeleteRoute(int(fd), rtentry(ip))
}

func rtentry(ip net.IP) *RTEntry {
	return &RTEntry{
		Dst: unix.RawSockaddrInet4{
			Family: unix.AF_INET,
			Addr:   [...]byte{ip[12], ip[13], ip[14], ip[15]},
		},
		Genmask: unix.RawSockaddrInet4{
			Family: unix.AF_INET,
			Addr:   [...]byte{0xff, 0xff, 0xff, 0xff},
		},
		Gateway: unix.RawSockaddrInet4{
			Family: unix.AF_INET,
		},
		Flags: unix.RTF_UP | unix.RTF_HOST | unix.RTF_REJECT,
	}
}

type RTEntry struct {
	Pad1    uint64
	Dst     unix.RawSockaddrInet4
	Gateway unix.RawSockaddrInet4
	Genmask unix.RawSockaddrInet4
	Flags   uint16
	Pad2    int16
	Pad3    uint64
	Tos     uint8
	Class   uint8
	Pad4    [3]int16
	Metric  int16
	Dev     *int8
	Mtu     uint64
	Window  uint64
	Irtt    uint16
	Pad5    [6]byte
}

func IoctlAddRoute(destFd int, route *RTEntry) error {
	return ioctlPtr(destFd, unix.SIOCADDRT, unsafe.Pointer(route))
}

func IoctlDeleteRoute(destFd int, route *RTEntry) error {
	return ioctlPtr(destFd, unix.SIOCDELRT, unsafe.Pointer(route))
}

// These helpers explicitly copy the contents of in into out to produce
// the correct sockaddr structure, without relying on unsafe casting to
// a type of a larger size.
func SockaddrInet4ToAny(in unix.RawSockaddrInet4) unix.RawSockaddrAny {
	var out unix.RawSockaddrAny
	copy(
		(*(*[unix.SizeofSockaddrAny]byte)(unsafe.Pointer(&out)))[:],
		(*(*[unix.SizeofSockaddrInet4]byte)(unsafe.Pointer(&in)))[:],
	)
	return out
}

func RawSockaddrInet4(rsa *unix.RawSockaddrInet4) *unix.RawSockaddrInet4 {
	return (*unix.RawSockaddrInet4)(unsafe.Pointer(rsa))
}

func ioctlPtr(fd int, req uint, arg unsafe.Pointer) (err error) {
	_, _, e1 := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(req), uintptr(arg))
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

// Do the interface allocations only once for common
// Errno values.
var (
	errEAGAIN error = unix.EAGAIN
	errEINVAL error = unix.EINVAL
	errENOENT error = unix.ENOENT
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e unix.Errno) error {
	switch e {
	case 0:
		return nil
	case unix.EAGAIN:
		return errEAGAIN
	case unix.EINVAL:
		return errEINVAL
	case unix.ENOENT:
		return errENOENT
	}
	return e
}
