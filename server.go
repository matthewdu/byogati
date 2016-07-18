package byogati

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/matthewdu/base62"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
)

const (
	domain          string = "abrv.in"
	gaUrl           string = "https://www.google-analytics.com/collect"
	gaDebugUrl      string = "https://www.google-analytics.com/debug/collect"
	reCaptchaUrl    string = "https://www.google.com/recaptcha/api/siteverify"
	reCaptchaSecret string = "6LdAUCUTAAAAAG16sFm3arMt2MQwEmaNnUH7UJ3Q"
)

var gaPost = delay.Func("gaPost", func(ctx context.Context, m url.Values, path string) {
	client := urlfetch.Client(ctx)
	resp, err := client.PostForm(gaUrl, m)
	if err != nil {
		log.Errorf(ctx, "payload failed to send: _%s", err)
	}
	log.Infof(ctx, "payload sent: _path=%s _payload=%s _status=%s", path, m.Encode(), resp.Status)
	//s, _ := ioutil.ReadAll(resp.Body)
	//log.Infof(ctx, "%s", string(s))
})

type ReCaptchaResponse struct {
	Success bool `json:"success"`
}

func init() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/s/:base62-id", redirect)
	r.GET("/debug/:base62-id", debug)
	r.GET("/l/:params", redirectWithParams)
	r.POST("/create", create)

	http.Handle("/", r)
}

func redirect(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	base62Id := c.Param("base62-id")
	var link Link
	if err := getLinkFromDatastore(ctx, base62Id, &link); err != nil {
		c.Status(400)
		log.Infof(ctx, "/s/%s: %s", base62Id, err)
		return
	}

	defer func() {
		payloadValues, _ := url.ParseQuery(link.Payload)
		enrichPayload(&payloadValues, c)
		task, _ := gaPost.Task(payloadValues, c.Request.URL.Path)
		if _, err := taskqueue.Add(ctx, task, ""); err != nil {
			log.Errorf(ctx, "/s/%s: %s", base62Id, err)
		}
	}()

	c.Redirect(301, link.Url)
}

func debug(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	base62Id := c.Param("base62-id")
	var link Link
	getLinkFromDatastore(ctx, base62Id, &link)

	c.JSON(200, gin.H{
		"url":     link.Url,
		"payload": link.Payload,
	})
}

func redirectWithParams(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	params := c.Param("params")

	payloadValues, _ := url.ParseQuery(params)
	url, _ := url.QueryUnescape(payloadValues.Get("url"))
	defer func() {
		payloadValues.Del("url")
		enrichPayload(&payloadValues, c)
		task, _ := gaPost.Task(payloadValues, c.Request.URL.Path)
		if _, err := taskqueue.Add(ctx, task, ""); err != nil {
			log.Errorf(ctx, "/l/%s: %s", params, err)
		}
	}()

	c.Redirect(301, url)
}

func create(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	// Verify reCaptcha
	reCaptchaRequest := url.Values{
		"secret":   []string{reCaptchaSecret},
		"response": []string{c.Request.FormValue("g-recaptcha-response")},
	}
	client := urlfetch.Client(ctx)
	resp, err := client.PostForm(reCaptchaUrl, reCaptchaRequest)
	if err != nil {
		log.Errorf(ctx, "/create: failed to verify reCaptcha %s", err)
	}
	s, _ := ioutil.ReadAll(resp.Body)
	var reCaptchaResp ReCaptchaResponse
	if err = json.Unmarshal(s, &reCaptchaResp); err != nil {
		log.Errorf(ctx, "/create: failed to parse reCaptchaResponse %s", err)
	}
	if !reCaptchaResp.Success {
		log.Infof(ctx, "/create: failed reCaptcha verification")
		c.String(400, "Failed reCaptcha verification.")
		return
	}

	link, err := makeLink(c)
	if err != nil {
		log.Infof(ctx, "/create: %s", err)
		c.String(400, err.Error())
		return
	}

	parsedUrl, _ := url.Parse(link.Url)
	if !parsedUrl.IsAbs() || parsedUrl.Host == domain {
		c.String(400, "Invalid URL")
		return
	}

	k := datastore.NewIncompleteKey(ctx, "link", nil)

	var key *datastore.Key
	key, err = datastore.Put(ctx, k, &link)
	if err != nil {
		log.Errorf(ctx, "/create: %s", err)
		c.String(500, "Server error, try again later")
		return
	}
	base62Id := base62.Encode(uint64(key.IntID()))
	log.Infof(ctx, "/create: Created _id=%d /s/%s", key.IntID(), base62Id)
	c.JSON(201, gin.H{
		"shortLink": "/s/" + base62Id,
		//"longLink":  "/l/url=" + url.QueryEscape(link.Url) + "&" + link.Payload,
	})
}

func enrichPayload(m *url.Values, c *gin.Context) {
	m.Set("ua", c.Request.UserAgent())
	m.Set("uip", c.ClientIP())
	m.Set("ds", "abrv")
	m.Set("cid", uuid.New())
}
