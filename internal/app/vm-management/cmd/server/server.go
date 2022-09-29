package server

import (
	"log"

	"github.com/labstack/echo"
	"github.com/shirinebadi/vm-management-server/internal/app/vm-management/config"
	"github.com/shirinebadi/vm-management-server/internal/app/vm-management/data/db"
	"github.com/shirinebadi/vm-management-server/internal/app/vm-management/handler"
	"github.com/spf13/cobra"
)

func main(cfg config.Config) {
	myDB, err := db.Init()
	if err != nil {
		log.Fatal("Failed to setup db: ", err.Error())
	}

	e := echo.New()

	userI := db.Mydb{DB: myDB}
	token := handler.Token{Cfg: cfg}

	user := handler.UserHandler{UserI: &userI, Token: token}
	mainHandler := handler.MainHandler{}

	e.POST("/register", user.Register)
	e.POST("/login", user.Login)
	e.POST("/main", mainHandler.ReadJson)

	address := cfg.Server.Address

	if err := e.Start(address); err != nil {
		log.Fatal(err)
	}
}

func Register(root *cobra.Command, cfg config.Config) {
	runServer := &cobra.Command{
		Use:   "server",
		Short: "server for virtual box management",
		Run: func(cmd *cobra.Command, args []string) {
			main(cfg)
		},
	}

	root.AddCommand(runServer)
}
