package poweradapter

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

const (
	Domain                 string = "example.org"
	MeasurementProtocolUrl string = "https://www.google-analytics.com/collect"
)

func init() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/r/:base-36-id", redirect)
	r.GET("/new", new)
	r.POST("/create", create)

	http.Handle("/", r)
}

func redirect(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	strId := c.Param("base-36-id")
	id, err := strconv.ParseInt(strId, 36, 64)
	if err != nil {
		c.String(400, err.Error())
		return
	}
	key := datastore.NewKey(ctx, "link", "", id, nil)

	var link Link
	if err = datastore.Get(ctx, key, &link); err != nil {
		c.String(400, err.Error())
		return
	}

	client := urlfetch.Client(ctx)
	var resp *http.Response
	resp, err = client.Post(MeasurementProtocolUrl, "", bytes.NewBufferString(link.Payload))
	if err != nil {
		log.Errorf(ctx, "%s", err)
	}
	log.Infof(ctx, "%s", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Infof(ctx, "%s", string(body))

	c.Redirect(301, link.Url)
}

func new(c *gin.Context) {
	c.HTML(200, "new.tmpl", nil)
}

func create(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	link, err := makeLink(c)
	if err != nil {
		c.String(400, err.Error())
		return
	}

	k := datastore.NewIncompleteKey(ctx, "link", nil)

	var key *datastore.Key
	key, err = datastore.Put(ctx, k, &link)
	if err != nil {
		c.String(400, err.Error())
		return
	}
	c.JSON(201, gin.H{
		"base 36 key.IntID()": Domain + "/" + strconv.FormatInt(key.IntID(), 36),
	})
}

func makeLink(c *gin.Context) (Link, error) {
	var link Link
	if err := c.Bind(&link); err != nil {
		return link, err
	}
	// TODO: Parameter checking on Link so that a HitType contains all its required parameters
	return link, nil
}
