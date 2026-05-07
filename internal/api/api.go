package api

import (
	"log"

	"github.com/gin-gonic/gin"
)

type WebAPI struct {
	ginEngine *gin.Engine
}

func NewAPI() *WebAPI {
	engine := gin.Default()

	return &WebAPI{
		ginEngine: engine,
	}
}

func (api *WebAPI) Run() {
	log.Fatal(api.ginEngine.Run(":8080"))
}