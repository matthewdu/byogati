package poweradapter

import (
	"base62"
	"bytes"
	//	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
)

const (
	domain string = "example.org"
	gaUrl  string = "https://www.google-analytics.com/collect"
)

var gaPost = delay.Func("gaPost", func(ctx context.Context, m url.Values, base62Id string, id int64) {
	client := urlfetch.Client(ctx)
	resp, err := client.Post(gaUrl, "", bytes.NewBufferString(m.Encode()))
	if err != nil {
		log.Errorf(ctx, "/r/%s: _id=%d %s", base62Id, id, err)
	}
	log.Infof(ctx, "/r/%s: _id=%d _payload=%s _status=%s", base62Id, id, m.Encode(), resp.Status)
	//s, _ := ioutil.ReadAll(resp.Body)
	//log.Infof(ctx, "%s", string(s))
})

func init() {
	r := gin.Default()
	r.GET("/r/:base62-id", redirect)
	r.POST("/create", create)

	http.Handle("/", r)
}

func redirect(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	strId := c.Param("base62-id")
	id, err := base62.Decode(strId)
	if err != nil {
		log.Errorf(ctx, "/r/%s: _id=%d %s", strId, id, err)
		c.Status(400)
		return
	}
	key := datastore.NewKey(ctx, "link", "", id, nil)

	var link Link
	if err = datastore.Get(ctx, key, &link); err != nil {
		log.Errorf(ctx, "/r/%s: _id=%d %s", strId, id, err)
		c.Status(500)
		return
	}

	payloadValues, _ := url.ParseQuery(link.Payload)
	payloadValues.Set("ua", c.Request.UserAgent())
	task, _ := gaPost.Task(payloadValues, strId, id)
	defer func() {
		if _, err = taskqueue.Add(ctx, task, ""); err != nil {
			log.Errorf(ctx, "/r/%s: _id=%d %s", strId, id, err)
		}
	}()

	c.Redirect(301, link.Url)
}

func create(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	link, err := makeLink(c)
	if err != nil {
		log.Errorf(ctx, "/create: %s", err)
		c.Status(400)
		return
	}

	k := datastore.NewIncompleteKey(ctx, "link", nil)

	var key *datastore.Key
	key, err = datastore.Put(ctx, k, &link)
	if err != nil {
		log.Errorf(ctx, "/create: %s", err)
		c.Status(500)
		return
	}
	strId := base62.Encode(key.IntID())
	log.Infof(ctx, "/create: _id=%d Created /r/%s", key.IntID(), strId)
	c.JSON(201, gin.H{
		"link": "/r/" + base62.Encode(key.IntID()),
	})
}
