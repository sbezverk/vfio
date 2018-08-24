package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	vfio "github.com/sbezverk/vfio/vfio-utils"
)

const (
	nsmVFPrefix = "NSM_VFS_"
)

// vfioConfig is stuct used to store vfio device specific information
type vfioConfig struct {
	VFIODevice string `yaml:"vfioDevice" json:"vfioDevice"`
	PCIAddr    string `yaml:"pciAddr" json:"pciAddr"`
}

func getNetworkServicesConfigs() (map[string][]*vfioConfig, error) {
	found := false
	networkServices := map[string][]*vfioConfig{}
	for _, env := range os.Environ() {
		key := strings.Split(env, "=")[0]
		val := strings.Split(env, "=")[1]
		if strings.HasPrefix(key, nsmVFPrefix) {
			found = true
			networkServiceName := strings.Split(key, nsmVFPrefix)[1]
			f, err := os.Open(val)
			if err != nil {
				return nil, fmt.Errorf("failed to open config file for network service %s with error: %+v", networkServiceName, err)
			}
			defer f.Close()
			d := json.NewDecoder(f)
			vc := []*vfioConfig{}
			if err := d.Decode(&vc); err != nil {
				return nil, fmt.Errorf("failed to decode config file for network service %s with error: %+v", networkServiceName, err)
			}
			networkServices[networkServiceName] = vc
		}
	}
	if !found {
		return nil, fmt.Errorf("no any Network Services found in Environment Variables")
	}

	return networkServices, nil
}

func main() {
	groupStatus := vfio.GroupStatus{
		Argsz: uint32(unsafe.Sizeof(vfio.GroupStatus{})),
	}
	networkServices, err := getNetworkServicesConfigs()
	if err != nil {
		fmt.Printf("Something happened while getting network services config: %+v\n", err)
	} else {
		for k, v := range networkServices {
			fmt.Printf("Network Services: %s\n", k)
			for _, p := range v {
				fmt.Printf("\t device: %s pci address: %s\n", p.VFIODevice, p.PCIAddr)
			}
		}
	}

	// Attempting to open /dev/vfio/vfio
	container, err := syscall.Open("/dev/vfio/vfio", syscall.O_RDWR, 0777)
	if err != nil {
		fmt.Printf("Something happened while opening /dev/vfio/vfio, error: %+v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Open container succeeded, handle: %d\n", container)

	// Attempting to open group /dev/vfio/{group-id}
	// Since it is just an example, vlan10 network service is used
	v := networkServices["vlan10"]
	group, err := syscall.Open(v[0].VFIODevice, syscall.O_RDWR, 0777)
	if err != nil {
		fmt.Printf("Something happened while opening %s, error: %+v\n", v[0].VFIODevice, err)
		os.Exit(1)
	}
	fmt.Printf("Open group succeeded, handle: %d\n", group)
	if err := vfio.GetGroupStatus(group, &groupStatus); err != nil {
		fmt.Printf("Fail to get group status with error: %+v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Group %d status Flags: %b \n", group, groupStatus.Flags)
	if (groupStatus.Flags & vfio.VFIO_GROUP_FLAGS_VIABLE) != vfio.VFIO_GROUP_FLAGS_VIABLE {
		fmt.Printf("The group is not viable, exiting...\n")
		os.Exit(1)
	}

	found, err := vfio.CheckExtension(container, vfio.VFIO_TYPE1_IOMMU)
	if err != nil {
		fmt.Printf("Failed to check for supported extension: %04x with error: %+v\n", vfio.VFIO_TYPE1_IOMMU, err)
		os.Exit(1)
	}
	if found {
		fmt.Printf("Device: %d supports VFIO_TYPE1_IOMMU\n", container)
	} else {
		fmt.Printf("Device: %d does not support VFIO_TYPE1_IOMMU\n", container)
	}

	found, err = vfio.CheckExtension(container, vfio.VFIO_NOIOMMU_IOMMU)
	if err != nil {
		fmt.Printf("Failed to check for supported extension: %04x with error: %+v\n", vfio.VFIO_NOIOMMU_IOMMU, err)
		os.Exit(1)
	}
	if found {
		fmt.Printf("Device: %d supports VFIO_NOIOMMU_IOMMU\n", container)
	} else {
		fmt.Printf("Device: %d does not support VFIO_NOIOMMU_IOMMU\n", container)
	}

	if err := vfio.SetGroupContainer(group, container); err != nil {
		fmt.Printf("Fail to set group's container with error: %+v\n", err)
		os.Exit(1)
	}
	if err := vfio.GetGroupStatus(group, &groupStatus); err != nil {
		fmt.Printf("Fail to get group status with error: %+v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Group %d status: %+v Flags: %b \n", group, groupStatus, groupStatus.Flags)
	fmt.Printf("PCI Address: %s\n", v[0].PCIAddr)
	pciAddr := v[0].PCIAddr
	device, err := vfio.GetGroupFD(group, pciAddr)
	if err != nil {
		fmt.Printf("Fail to get group file descriptor %+v.\n", err)
		os.Exit(1)
	}
	fmt.Printf("Group %d file descriptor is %d\n", group, device)
	deviceInfo := vfio.DeviceInfo{
		Argsz: uint32(unsafe.Sizeof(vfio.DeviceInfo{})),
	}
	if err := vfio.GetDeviceInfo(device, &deviceInfo); err != nil {
		fmt.Printf("Fail to get device info with error: %+v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Group %d device info: %+v\n", group, deviceInfo)
}
