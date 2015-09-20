package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"path/filepath"
	"strings"
	"sync"

	"github.com/appropriate/go-virtualboxclient/vboxwebsrv"
	"github.com/appropriate/go-virtualboxclient/virtualboxclient"
	"github.com/calavera/dkvolume"
)

const (
	pluginId = "virtualbox"

	defaultVboxwebsrvUrl = "http://192.168.99.1:18083"
)

var (
	socketAddress       = filepath.Join("/run/docker/plugins", fmt.Sprintf("%s.sock", pluginId))
	defaultMountRoot    = filepath.Join(dkvolume.DefaultDockerRootDirectory, fmt.Sprintf("_%s", pluginId))
	mountRoot           = flag.String("mount-root", defaultMountRoot, "VirtualBox volumes root mount directory (on VM)")
	storageLocationRoot = flag.String("storage-location-root", "", "VirtualBox volumes root storage location directory (on host)")

	vboxwebsrvUsername = flag.String("vboxwebsrv-username", "", "Username to connect to vboxwebsrv")
	vboxwebsrvPassword = flag.String("vboxwebsrv-password", "", "Password to connect to vboxwebsrv")
	vboxwebsrvUrl      = flag.String("vboxwebsrv-url", defaultVboxwebsrvUrl, "URL to connect to vboxwebsrv")
)

type virtualboxDriver struct {
	sync.Mutex

	virtualbox *virtualboxclient.VirtualBox
	machine    *virtualboxclient.Machine
	volumes    map[string]*virtualboxclient.Medium
}

func (d virtualboxDriver) Create(r dkvolume.Request) dkvolume.Response {
	d.Lock()
	defer d.Unlock()

	var err error

	fmt.Printf("Creating volume %#v\n", r)

	var medium *virtualboxclient.Medium
	if medium, err = d.virtualbox.CreateHardDisk("vmdk", d.storageLocation(r.Name)); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	fmt.Printf("Hard disk: %#v\n", medium)

	if _, err = medium.CreateBaseStorage(1000000, nil); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	d.volumes[r.Name] = medium

	return dkvolume.Response{}
}

func (d virtualboxDriver) Mount(r dkvolume.Request) dkvolume.Response {
	fmt.Printf("Mounting volume %#v\n", r)

	scs, err := d.machine.GetStorageControllers()
	if err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	for _, sc := range scs {
		fmt.Printf("%#v\n", sc)
	}

	return dkvolume.Response{}
}

func (d virtualboxDriver) Path(r dkvolume.Request) dkvolume.Response {
	mountpoint := d.mountPoint(r.Name)

	fmt.Printf("Path %#v => %s\n", r, mountpoint)

	return dkvolume.Response{Mountpoint: d.mountPoint(r.Name)}
}

func (d virtualboxDriver) Remove(r dkvolume.Request) dkvolume.Response {
	d.Lock()
	defer d.Unlock()

	fmt.Printf("Removing volume %#v\n", r)

	if m, ok := d.volumes[r.Name]; ok {
		if _, err := m.DeleteStorage(); err != nil {
			return dkvolume.Response{Err: err.Error()}
		}
	}

	return dkvolume.Response{}
}

func (d virtualboxDriver) Unmount(r dkvolume.Request) dkvolume.Response {
	fmt.Printf("Unmounting volume %#v\n", r)
	return dkvolume.Response{}
}

func (d virtualboxDriver) storageLocation(name string) string {
	return filepath.Join(*storageLocationRoot, fmt.Sprintf("%s.vmdk", name))
}

func (d virtualboxDriver) mountPoint(name string) string {
	return filepath.Join(*mountRoot, name)
}

func (d virtualboxDriver) findCurrentMachine() (*virtualboxclient.Machine, error) {
	var (
		sp                 *virtualboxclient.SystemProperties
		machines           []*virtualboxclient.Machine
		chipset            *vboxwebsrv.ChipsetType
		maxNetworkAdapters uint32
		na                 *virtualboxclient.NetworkAdapter
		mac                string
		intfMacs           []string
		interfaces         []net.Interface
		err                error
	)

	interfaces, err = net.Interfaces()
	if err != nil {
		return nil, err
	}

	intfMacs = make([]string, len(interfaces))
	for i, intf := range interfaces {
		intfMacs[i] = strings.ToUpper(strings.Replace(intf.HardwareAddr.String(), ":", "", -1))
	}

	sp, err = d.virtualbox.GetSystemProperties()
	if err != nil {
		return nil, err
	}

	machines, err = d.virtualbox.GetMachines()
	if err != nil {
		return nil, err
	}

	for _, m := range machines {
		chipset, err = m.GetChipsetType()
		if err != nil {
			return nil, err
		}

		maxNetworkAdapters, err = sp.GetMaxNetworkAdapters(chipset)
		if err != nil {
			return nil, err
		}

		for i := uint32(0); i < maxNetworkAdapters; i++ {
			na, err = m.GetNetworkAdapter(i)
			if err != nil {
				return nil, err
			}

			mac, err = na.GetMACAddress()
			if err != nil {
				return nil, err
			}

			for _, intfMac := range intfMacs {
				if mac == intfMac {
					return m, nil
				}
			}
		}
	}

	return nil, errors.New("Unable to find machine")
}

// TODO: Figure out why Ctrl+C doesn't immediately terminate
func main() {
	flag.Parse()

	if storageLocationRoot == nil || *storageLocationRoot == "" {
		log.Fatal("Storage location root must be specified with -storage-location-root")
	}

	virtualbox := virtualboxclient.New(*vboxwebsrvUsername, *vboxwebsrvPassword, *vboxwebsrvUrl)

	if err := virtualbox.Logon(); err != nil {
		log.Fatal(err)
	}

	d := virtualboxDriver{
		virtualbox: virtualbox,
		volumes:    map[string]*virtualboxclient.Medium{},
	}

	if m, err := d.findCurrentMachine(); err != nil {
		log.Fatal(err)
	} else {
		d.machine = m
	}

	h := dkvolume.NewHandler(d)
	fmt.Printf("Listening on %s\n", socketAddress)
	fmt.Println(h.ServeUnix("root", socketAddress))
}
