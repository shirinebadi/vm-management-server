package handler

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"strconv"

	"github.com/labstack/echo"
	vbox "github.com/pyToshka/go-virtualbox"
	"github.com/shirinebadi/vm-management-server/internal/app/vm-management/request"
	"golang.org/x/crypto/ssh"
)

func getKeyFile() (key ssh.Signer, err error) {
	usr, _ := user.Current()
	file := usr.HomeDir + "/.ssh/id_rsa"
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	key, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		return
	}

	return
}

func sshClient() {
	key, err := getKeyFile()
	if err != nil {
		panic(err)
	}

	config := &ssh.ClientConfig{
		User: "shirinei",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	client, err := ssh.Dial("tcp", "192.168.56.101:22", config)

	session, err := client.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()
	sessStdOut, err := session.StdoutPipe()
	if err != nil {
		panic(err)
	}
	go io.Copy(os.Stdout, sessStdOut)
	sessStderr, err := session.StderrPipe()
	if err != nil {
		panic(err)
	}
	go io.Copy(os.Stderr, sessStderr)
	err = session.Run("/user/bin/whoami") // eg., /usr/bin/whoami
	if err != nil {
		panic(err)
	}

}

type MainHandler struct{}

func (m *MainHandler) ReadJson(c echo.Context) error {
	req := request.Request{}
	defer c.Request().Body.Close()
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Failed to read the boy: %s", err)
		return c.String(http.StatusInternalServerError, "")
	}

	err = json.Unmarshal(b, &req)
	log.Printf("This is your request: %v", req)

	switch req.Command {
	case "status":

		if req.VmName != "" {
			vm, _ := vbox.GetMachine(req.VmName)
			log.Println(vm.State)
			return c.JSON(http.StatusOK, map[string]string{
				"command": "status",
				"VmName":  req.VmName,
				"Status":  string(vm.State),
			})
		}
		if req.VmName == "" {

			machines, _ := vbox.ListMachines()
			size := len(machines)
			arr := [100]map[string]string{}
			for v := 0; v < len(machines); v++ {

				arr[v] = map[string]string{

					"VmName": machines[v].Name,
					"Status": string(machines[v].State),
				}
			}
			log.Println(len(machines), arr[:size])
			return c.JSON(http.StatusOK, arr[:size])
		}
	case "on/off":
		if req.VmName == "" {
			return c.String(http.StatusBadRequest, "Tell Me The Name of VM")
		}
		vm, _ := vbox.GetMachine(req.VmName)
		if vm.State != "running" {
			vm.Start()
			return c.JSON(http.StatusOK, map[string]string{
				"command": "on/of",
				"VmName":  req.VmName,
				"Status":  "Powered On",
			})
		}

		if vm.State != "poweroff" {
			vm.Poweroff()
			return c.JSON(http.StatusOK, map[string]string{
				"command": "on/off",
				"VmName":  req.VmName,
				"Status":  "Powered Off",
			})
		}
	case "settings":
		if req.VmName == "" {
			return c.String(http.StatusBadRequest, "Tell Me The Name of VM")
		}
		vm, _ := vbox.GetMachine(req.VmName)
		if vm.State == "running" {
			return c.String(http.StatusBadRequest, "Can't Change Settings While Running")
		}
		vm.CPUs = req.Cpu
		vm.Memory = req.Ram
		err := vm.Modify()
		if err != nil {
			return c.String(http.StatusInternalServerError, "Unable to Modify")
		}

		log.Println(vm.VRAM, vm.CPUs)
		return c.JSON(http.StatusOK, map[string]string{
			"command": "Settings",
			"VmName":  req.VmName,
			"CPU":     strconv.FormatUint(uint64(vm.CPUs), 10),
			"Ram":     strconv.FormatUint(uint64(vm.Memory), 10),
			"Status":  "Done",
		})
	case "clone":
		if req.SourceVmName == "" || req.DestVmName == "" {
			return c.String(http.StatusBadRequest, "Give Me more Details")
		}
		vm, _ := vbox.GetMachine(req.SourceVmName)
		vmNew, _ := vbox.CreateMachine(req.DestVmName, "")
		vmNew.CPUs = vm.CPUs
		vmNew.VRAM = vm.VRAM
		vmNew.BootOrder = vm.BootOrder
		vmNew.Memory = vm.Memory
		vmNew.Flag = vm.Flag
		vmNew.OSType = vm.OSType
		log.Println(vm)
		log.Println("and\n", vmNew)
		return c.JSON(http.StatusOK, map[string]string{
			"command":      "Clone",
			"SourceVMName": req.SourceVmName,
			"DestVmName":   req.DestVmName,
			"Status":       "Done",
		})

	case "delete":
		if req.VmName == "" {
			return c.String(http.StatusBadRequest, "Tell Me The Name of VM")
		}
		vm, _ := vbox.GetMachine(req.VmName)
		vm.Delete()
		return c.JSON(http.StatusOK, map[string]string{
			"command": "Delete",
			"VmName":  req.VmName,
			"Status":  "Done",
		})
	case "execute":
		sshClient()

	default:
		return c.String(http.StatusOK, "Tell me valid Command")

	}

	return c.String(http.StatusOK, "")

}
