package byogati

import (
	"github.com/gin-gonic/gin"
	"github.com/matthewdu/base62"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type Link struct {
	Url     string `json:"url" form:"url" binding:"required"`
	Payload string `json:"payload" form:"payload" binding:"required"`
}

func makeLink(c *gin.Context) (link Link, err error) {
	if err = c.Bind(&link); err != nil {
		return link, err
	}
	return link, nil
}

func getLinkFromDatastore(ctx context.Context, base62Id string, link *Link) error {
	uid, err := base62.Decode(base62Id)
	if err != nil {
		return err
	}
	id := int64(uid)
	key := datastore.NewKey(ctx, "link", "", id, nil)

	if err = datastore.Get(ctx, key, link); err != nil {
		return err
	}
	return nil
}
