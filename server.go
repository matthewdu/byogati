package poweradapter

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

const (
	Domain string = "example.org"
)

func init() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	//r.GET("/test-create", testCreateLink)
	//r.GET("/test-load/:base-36-id", testLoadLink)
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

	// TODO: Send payload to Google Analytics
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

func testLoadLink(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	strId := c.Param("base-36-id")
	id, _ := strconv.ParseInt(strId, 36, 64)
	k := datastore.NewKey(ctx, "link", "", id, nil)

	var l Link
	if err := datastore.Get(ctx, k, &l); err != nil {
		c.String(500, k.Encode())
		c.String(500, err.Error())
	}
	c.String(200, l.Payload)
	c.JSON(200, gin.H{
		"l.TrackingId": l.TrackingId,
		"l.HitType":    l.HitType,
		"l.Payload":    l.Payload,
	})
}

func testCreateLink(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	l := &Link{
		Url:           "http://matthewdu.com",
		TrackingId:    "TestTrackingId",
		HitType:       "event",
		EventCategory: "TestCat",
		EventAction:   "TestAct",
		EventLabel:    "TestLabel",
		EventValue:    4,
	}

	k := datastore.NewIncompleteKey(ctx, "link", nil)
	key, err := datastore.Put(ctx, k, l)
	if err != nil {
		panic(err)
		fmt.Println(err)
	}
	c.JSON(200, gin.H{
		"key.Encode()":        key.Encode(),
		"key.IntID()":         key.IntID(),
		"base 36 key.IntID()": strconv.FormatInt(key.IntID(), 36),
	})

	//fmt.Println(k.Encode())
}
