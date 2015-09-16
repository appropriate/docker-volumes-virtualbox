package main

import (
	"fmt"
	"path/filepath"

	"github.com/appropriate/go-virtualboxclient/vboxwebsrv"
	"github.com/calavera/dkvolume"
)

type vboxwebsrvDriver struct {
}

func (g vboxwebsrvDriver) Create(r dkvolume.Request) dkvolume.Response {
	fmt.Printf("Creating volume %#v\n", r)

	svc := vboxwebsrv.NewVboxPortType("http://192.168.99.1:18083", false, nil)

	fmt.Printf("Logging on %#v\n", svc)
	vbox, err := svc.IWebsessionManagerlogon(&vboxwebsrv.IWebsessionManagerlogon{})
	if err != nil {
		return dkvolume.Response{Err: err.Error()}
	}
	fmt.Printf("Response: %#v\n", vbox)

	location := filepath.Join("/Users/mike/.docker/machine/machines", fmt.Sprintf("%s.vmdk"))
	request := vboxwebsrv.IVirtualBoxcreateHardDisk{This: vbox.Returnval, Format: "vmdk", Location: location}
	response, err := svc.IVirtualBoxcreateHardDisk(&request)
	if err != nil {
		return dkvolume.Response{Err: err.Error()}
	}

	fmt.Printf("Response: %#v\n", response)

	return dkvolume.Response{}
}

func (g vboxwebsrvDriver) Mount(r dkvolume.Request) dkvolume.Response {
	fmt.Printf("Mounting volume %#v\n", r)
	return dkvolume.Response{}
}

func (g vboxwebsrvDriver) Path(r dkvolume.Request) dkvolume.Response {
	fmt.Printf("Pathing volume %#v\n", r)
	return dkvolume.Response{}
}

func (g vboxwebsrvDriver) Remove(r dkvolume.Request) dkvolume.Response {
	fmt.Printf("Removing volume %#v\n", r)
	return dkvolume.Response{}
}

func (g vboxwebsrvDriver) Unmount(r dkvolume.Request) dkvolume.Response {
	fmt.Printf("Unmounting volume %#v\n", r)
	return dkvolume.Response{}
}

func main() {
	d := vboxwebsrvDriver{}
	h := dkvolume.NewHandler(d)
	fmt.Println(h.ServeUnix("root", "vboxwebsrv"))
}
