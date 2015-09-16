package main

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"github.com/appropriate/go-virtualboxclient/virtualboxclient"
	"github.com/calavera/dkvolume"
)

type virtualboxDriver struct {
	sync.Mutex

	client  *virtualboxclient.VirtualBoxClient
	volumes map[string]*virtualboxclient.Medium
}

func (d virtualboxDriver) Create(r dkvolume.Request) dkvolume.Response {
	d.Lock()
	defer d.Unlock()

	var err error

	fmt.Printf("Creating volume %#v\n", r)

	var medium *virtualboxclient.Medium
	if medium, err = d.client.CreateHardDisk("vmdk", d.storageLocation(r.Name)); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	fmt.Printf("Hard disk: %#v\n", medium)

	if err = medium.CreateBaseStorage(1000000, nil); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	d.volumes[r.Name] = medium

	return dkvolume.Response{}
}

func (d virtualboxDriver) Mount(r dkvolume.Request) dkvolume.Response {
	fmt.Printf("Mounting volume %#v\n", r)
	return dkvolume.Response{}
}

func (d virtualboxDriver) Path(r dkvolume.Request) dkvolume.Response {
	fmt.Printf("Pathing volume %#v\n", r)
	return dkvolume.Response{}
}

func (d virtualboxDriver) Remove(r dkvolume.Request) dkvolume.Response {
	d.Lock()
	defer d.Unlock()

	fmt.Printf("Removing volume %#v\n", r)

	if m, ok := d.volumes[r.Name]; ok {
		if err := m.DeleteStorage(); err != nil {
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
	return filepath.Join("/Users/mike/.docker/machine/machines/dev", fmt.Sprintf("%s.vmdk", name))
}

// TODO: Figure out why Ctrl+C doesn't immediately terminate
func main() {
	client := virtualboxclient.New("", "", "http://192.168.99.1:18083")

	if err := client.Logon(); err != nil {
		log.Fatal(err)
	}

	d := virtualboxDriver{
		client:  client,
		volumes: map[string]*virtualboxclient.Medium{},
	}
	h := dkvolume.NewHandler(d)
	fmt.Println(h.ServeUnix("root", "virtualbox"))
}
