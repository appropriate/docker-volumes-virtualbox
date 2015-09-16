package main

import (
	"fmt"
	"path/filepath"

	"github.com/appropriate/go-virtualboxclient/virtualboxclient"
	"github.com/calavera/dkvolume"
)

type virtualboxDriver struct {
}

func (d virtualboxDriver) Create(r dkvolume.Request) dkvolume.Response {
	fmt.Printf("Creating volume %#v\n", r)

	client := virtualboxclient.New("", "", "http://192.168.99.1:18083")

	if err := client.Logon(); err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	location := filepath.Join("/Users/mike/.docker/machine/machines", fmt.Sprintf("%s.vmdk"))

	hardDisk, err := client.CreateHardDisk("vmdk", location)
	if err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	fmt.Printf("Hard disk: %#v\n", hardDisk)

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
	fmt.Printf("Removing volume %#v\n", r)
	return dkvolume.Response{}
}

func (d virtualboxDriver) Unmount(r dkvolume.Request) dkvolume.Response {
	fmt.Printf("Unmounting volume %#v\n", r)
	return dkvolume.Response{}
}

func main() {
	d := virtualboxDriver{}
	h := dkvolume.NewHandler(d)
	fmt.Println(h.ServeUnix("root", "vboxwebsrv"))
}
