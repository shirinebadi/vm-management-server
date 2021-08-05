package main

import (
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	vbox "github.com/pyToshka/go-virtualbox"
)


type JwtClaims struct{
	Name     string
	jwt.StandardClaims
}

func getKeyFile() (key ssh.Signer, err error){
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
	if  err !=nil {
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



func createJwtToken()(string,error){
	claim := JwtClaims{
		 "shirin",
		jwt.StandardClaims{
		 	Id: "main_user_id",
		 	ExpiresAt: time.Now().Add(24*time.Hour).Unix(),

		},
	}
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS512,claim)
	token, err := rawToken.SignedString([]byte("mySecret"))
	if err !=nil{
		return "", err
	}

	return token,nil
}

  func readJson(c echo.Context) error{
  	req := request{}
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
				"command" : "status",
				"VmName" : req.VmName,
				"Status" : string(vm.State),

			})
		}
		if req.VmName ==""{


			machines, _ := vbox.ListMachines()
			size := len(machines)
			arr := [100] map[string]string{}
			for v:=0;v< len(machines);v++{

				arr[v] = map [string]string{

					"VmName" : machines[v].Name,
					"Status" : string(machines[v].State),
				}
			}
			log.Println(len(machines), arr[:size])
			return c.JSON(http.StatusOK, arr[:size])
		}
	  case "on/off":
	  	if req.VmName ==""{
	  		return c.String(http.StatusBadRequest, "Tell Me The Name of VM")
		}
		  vm, _ := vbox.GetMachine(req.VmName)
		  if vm.State!="running" {
			  vm.Start()
			  return c.JSON(http.StatusOK, map[string]string{
				  "command" : "on/of",
				  "VmName" : req.VmName,
				  "Status" : "Powered On",

			  })
		  }

		  if vm.State!="poweroff" {
		  	vm.Poweroff()
			  return c.JSON(http.StatusOK, map[string]string{
				  "command" : "on/off",
				  "VmName" : req.VmName,
				  "Status" : "Powered Off",

			  })
		  }
	  case "settings":
		  if req.VmName ==""{
			  return c.String(http.StatusBadRequest, "Tell Me The Name of VM")
		  }
		  vm, _ := vbox.GetMachine(req.VmName)
		  if vm.State=="running" {
			  return c.String(http.StatusBadRequest, "Can't Change Settings While Running")
		  }
		  vm.CPUs = req.Cpu
		  vm.Memory = req.Ram
		  err := vm.Modify()
		  if err!=nil{
		  	return c.String(http.StatusInternalServerError, "Unable to Modify")
		  }

		  log.Println(vm.VRAM, vm.CPUs)
		  return c.JSON(http.StatusOK, map[string]string{
			  "command" : "Settings",
			  "VmName" : req.VmName,
			  "CPU" : strconv.FormatUint(uint64(vm.CPUs),10),
			  "Ram": strconv.FormatUint(uint64(vm.Memory),10),
			  "Status" : "Done",
		  })
	  case "clone":
	  	if req.SourceVmName == "" || req.DestVmName==""{
			return c.String(http.StatusBadRequest, "Give Me more Details")
		}
		vm, _ := vbox.GetMachine(req.SourceVmName)
		vmNew,_ := vbox.CreateMachine(req.DestVmName,"")
		vmNew.CPUs = vm.CPUs
		vmNew.VRAM = vm.VRAM
		vmNew.BootOrder = vm.BootOrder
		vmNew.Memory = vm.Memory
		vmNew.Flag = vm.Flag
		vmNew.OSType = vm.OSType
		log.Println(vm)
		log.Println("and\n", vmNew)
		  return c.JSON(http.StatusOK, map[string]string{
			  "command" : "Clone",
			  "SourceVMName" : req.SourceVmName,
			  "DestVmName": req.DestVmName,
			  "Status" : "Done",
		  })

	  case "delete":
		  if req.VmName ==""{
			  return c.String(http.StatusBadRequest, "Tell Me The Name of VM")
		  }
		  vm, _ := vbox.GetMachine(req.VmName)
		  vm.Delete()
		  return c.JSON(http.StatusOK, map[string]string{
			  "command" : "Delete",
			  "VmName" : req.VmName,
			  "Status" : "Done",

		  })
	  case "execute" :
		sshClient()





	  default:
	  	return c.String(http.StatusOK, "Tell me valid Command")

	  }


	  return c.String(http.StatusOK, "")

  }



  func login (c echo.Context)error{
  	username := c.QueryParam("username")
  	password := c.QueryParam("password")

  	if username == "shirinei" && password =="123456789"{
  		cookie := &http.Cookie{}

  		cookie.Name ="sessionPart"
  		cookie.Value="Some_string"
  		cookie.Expires = time.Now().Add(10 * time.Minute)
  		c.SetCookie(cookie)

  		token,err := createJwtToken()
  		if err!= nil{
  			log.Println("Token Error", err)
  			return c.String(http.StatusInternalServerError, "Error in token")
		}


  		return c.JSON(http.StatusOK, map[string]string{
  			"message": "Happily Logged in",
  			"token": token,
		})
	}

	return c.String(http.StatusInternalServerError, "Not Valid User and Pass")

  }



/*func ServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	token,_ := createJwtToken()
	return func(c echo.Context) error {
		c.Response().Header().Set("Token", token)


		return next(c)
	}
}
*/
  func main() {

  	counter := 0
	   e := echo.New()

	  // e.Use(ServerHeader)
	   mainGroup := e.Group("/login")
	  mainGroup.Use(middleware.BasicAuth(func(user, pass string, c echo.Context) (bool,error) {
	   if user =="shirinei" && pass =="123456789"{
	   	counter ++;
		   token, err := createJwtToken()
		   c.Response().Header().Set("Token", token)
		   if counter == 1 {
			   if err != nil {
				   log.Println("Token Error", err)
				   return false, c.String(http.StatusInternalServerError, "Error in token")
			   }
			   return true, c.JSON(http.StatusOK, map[string]string{
				   "message": "Happily Logged in",
				   "token":   token,
			   })

		   }
		   return true,nil;
	   }
	   return false,nil;
	   }))
	   //e.POST("/login", login)
	   //mainGroup.POST("/",login)
	  mainGroup.POST("/main", readJson)
	   e.Start(":8000")
	 }
