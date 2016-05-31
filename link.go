package poweradapter

import (
	"github.com/gin-gonic/gin"
)

type Link struct {
	Url     string `json:"url" form:"url" binding:"required"`
	Payload string `json:"payload" form:"payload" binding:"required"` //auto generated from PropertyLoadSaver.Load
}

func makeLink(c *gin.Context) (link Link, err error) {
	if err = c.Bind(&link); err != nil {
		return link, err
	}
	// TODO: Parameter checking on Link so that a HitType contains all its required parameters
	return link, nil
}
