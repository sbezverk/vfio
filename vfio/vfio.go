package vfio

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	_IOC_NRBITS   = 8
	_IOC_TYPEBITS = 8

	_IOC_SIZEBITS = 14
	_IOC_DIRBITS  = 2

	_IOC_NRMASK   = ((1 << _IOC_NRBITS) - 1)
	_IOC_TYPEMASK = ((1 << _IOC_TYPEBITS) - 1)
	_IOC_SIZEMASK = ((1 << _IOC_SIZEBITS) - 1)
	_IOC_DIRMASK  = ((1 << _IOC_DIRBITS) - 1)

	_IOC_NRSHIFT   = 0
	_IOC_TYPESHIFT = (_IOC_NRSHIFT + _IOC_NRBITS)
	_IOC_SIZESHIFT = (_IOC_TYPESHIFT + _IOC_TYPEBITS)
	_IOC_DIRSHIFT  = (_IOC_SIZESHIFT + _IOC_SIZEBITS)

	_IOC_NONE = uint8(0)

	_IOC_WRITE = uint8(1)

	_IOC_READ = uint8(2)

	VFIO_API_VERSION = 0

	VFIO_TYPE1_IOMMU   = 1
	VFIO_NOIOMMU_IOMMU = 8

	VFIO_TYPE = 0x3b
	VFIO_BASE = 100

	VFIO_GROUP_FLAGS_VIABLE        = uint32(1 << 0)
	VFIO_GROUP_FLAGS_CONTAINER_SET = uint32(1 << 1)

	VFIO_DEVICE_FLAGS_RESET = (1 << 0) /* Device supports reset */
	VFIO_DEVICE_FLAGS_PCI   = (1 << 1) /* vfio-pci device */

)

func VFIO_GROUP_GET_STATUS() uint32 {
	return _IO(VFIO_TYPE, VFIO_BASE+3)
}

func VFIO_CHECK_EXTENSION() uint32 {
	return _IO(VFIO_TYPE, VFIO_BASE+1)
}

func VFIO_GROUP_SET_CONTAINER() uint32 {
	return _IO(VFIO_TYPE, VFIO_BASE+4)
}

func VFIO_GROUP_GET_DEVICE_FD() uint32 {
	return _IO(VFIO_TYPE, VFIO_BASE+6)
}

func VFIO_DEVICE_GET_INFO() uint32 {
	return _IO(VFIO_TYPE, VFIO_BASE+7)
}

func _IOC(dir, iocType, nr, size uint32) uint32 {
	return (((dir) << _IOC_DIRSHIFT) |
		((iocType) << _IOC_TYPESHIFT) |
		((nr) << _IOC_NRSHIFT) |
		((size) << _IOC_SIZESHIFT))
}

func _IO(iocType, nr uint32) uint32 {
	return _IOC(uint32(_IOC_NONE), (iocType), (nr), 0)
}

func _IOC_TYPECHECK(t uint32) uint32 {
	return uint32((unsafe.Sizeof(t)))
}

func _IOR(iocType, nr, size uint32) uint32 {
	return _IOC(uint32(_IOC_READ), (iocType), (nr), (_IOC_TYPECHECK(size)))
}

func _IOW(iocType, nr, size uint32) uint32 {
	return _IOC(uint32(_IOC_WRITE), (iocType), (nr), (_IOC_TYPECHECK(size)))
}

func _IORW(iocType, nr, size uint32) uint32 {
	return _IOC(uint32(_IOC_READ|_IOC_WRITE), (iocType), (nr), (_IOC_TYPECHECK(size)))
}

func _IOR_BAD(iocType, nr, size uint32) uint32 {
	return _IOC(uint32(_IOC_READ), (iocType), (nr), uint32(unsafe.Sizeof(size)))
}

func _IOW_BAD(iocType, nr, size uint32) uint32 {
	return _IOC(uint32(_IOC_WRITE), (iocType), (nr), uint32(unsafe.Sizeof(size)))
}

func _IORW_BAD(iocType, nr, size uint32) uint32 {
	return _IOC(uint32(_IOC_READ|_IOC_WRITE), (iocType), (nr), uint32(unsafe.Sizeof(size)))
}

func _IOC_DIR(nr uint32) uint32 {
	return uint32(((nr) >> _IOC_DIRSHIFT) & _IOC_DIRMASK)
}

func _IOC_TYPE(nr uint32) uint32 {
	return uint32(((nr) >> _IOC_TYPESHIFT) & _IOC_TYPEMASK)
}

func _IOC_NR(nr uint32) uint32 {
	return uint32(((nr) >> _IOC_NRSHIFT) & _IOC_NRMASK)
}

func _IOC_SIZE(nr uint32) uint32 {
	return uint32(((nr) >> _IOC_SIZESHIFT) & _IOC_SIZEMASK)
}

func IOC_IN() uint32 {
	return uint32(_IOC_WRITE) << uint32(_IOC_DIRSHIFT)
}

func IOC_OUT() uint32 {
	return uint32(_IOC_READ) << uint32(_IOC_DIRSHIFT)
}

func IOC_INOUT() uint32 {
	return uint32(_IOC_WRITE|_IOC_READ) << uint32(_IOC_DIRSHIFT)
}

func IOCSIZE_MASK() uint32 {
	return uint32((_IOC_SIZEMASK) << uint32(_IOC_SIZESHIFT))
}

func IOCSIZE_SHIFT() uint32 {
	return uint32(_IOC_SIZESHIFT)
}

// GroupStatus used for VFIO_GROUP_GET_STATUS call
type GroupStatus struct {
	Argsz uint32
	Flags uint32
}

// DeviceInfo used to keep device related information
type DeviceInfo struct {
	Argsz      uint32
	Flags      uint32
	NumRegions uint32
	NumIRQs    uint32
}

// GetGroupStatus updates Flags field for a group specified by group parameter
func GetGroupStatus(group int, groupStatus *GroupStatus) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(group),
		uintptr(VFIO_GROUP_GET_STATUS()),
		uintptr(unsafe.Pointer(groupStatus)),
	)
	if errno != 0 {
		return fmt.Errorf("fail to get group status of %d with errno: %+v", group, errno)
	}
	return nil
}

// CheckExtension checks if parent vfio device supports specified extension
func CheckExtension(container int, extention uint32) (bool, error) {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(container),
		uintptr(VFIO_CHECK_EXTENSION()),
		uintptr(unsafe.Pointer(&extention)),
	)
	if errno != 0 {
		return false, fmt.Errorf("fail to check for extension for device %d with errno: %+v", container, errno)
	}
	return true, nil
}

// SetGroupContainer sets the container for the VFIO group to the open VFIO file
// descriptor provided.
func SetGroupContainer(group int, container int) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(group),
		uintptr(VFIO_GROUP_SET_CONTAINER()),
		uintptr(unsafe.Pointer(&container)),
	)
	if errno != 0 {
		return fmt.Errorf("fail to set container %d to the provided group %d with errno: %+v", container, group, errno)
	}
	return nil
}

// GetGroupFD gets File descriptor for a specified by PCI address device
func GetGroupFD(group int, pciDevice *string) (int, error) {
	fmt.Printf("VFIO_GROUP_GET_DEVICE_FD() returned: %04x\n", VFIO_GROUP_GET_DEVICE_FD())
	buffer := make([]byte, len(*pciDevice)+1)
	for i, c := range *pciDevice {
		buffer[i] = uint8(c)
	}
	buffer[len(*pciDevice)] = 0x0
	fmt.Printf("pciDevice: %s\n", string(buffer))
	device, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(group),
		uintptr(VFIO_GROUP_GET_DEVICE_FD()),
		uintptr(unsafe.Pointer(&buffer[0])),
	)
	if errno != 0 {
		return 0, fmt.Errorf("fail to get file descriptor for %d with errno: %+v", group, errno)
	}
	return int(device), nil
}

// GetDeviceInfo gets information about the specified device
func GetDeviceInfo(device int, deviceInfo *DeviceInfo) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(device),
		uintptr(VFIO_DEVICE_GET_INFO()),
		uintptr(unsafe.Pointer(deviceInfo)),
	)
	if errno != 0 {
		return fmt.Errorf("fail to get info for device fd %d with errno: %+v", device, errno)
	}
	return nil
}
