package poweradapter

import (
    "fmt"
    "net/http"

    "appengine"
    "appengine/datastore"
    "github.com/gin-gonic/gin"
)

func init() {
    r := gin.Default()
    r.GET("/ping", func(c *gin.Context) {
	c.JSON(200, gin.H{
	    "message": "pong",
	})
    })
    r.GET("/test-create", testCreateLink)

    http.Handle("/", r)
}

func testCreateLink(c *gin.Context) {
    ctx := appengine.NewContext(c.Request)
    
    eventHit := Event{
        EventCategory: "TestCat",
        EventAction: "TestAct",
        EventValue: 4,
    }

    l := &Link{
        Url: "http://matthewdu.com",
        TrackingId: "TestTrackingId",
        Hit: eventHit,
    }

    fmt.Println(l.PayloadString());

    k := datastore.NewIncompleteKey(ctx, "link", nil)
    _, err := datastore.Put(ctx, k, l)
    if err != nil {
	panic(err)
        fmt.Println(err)
    }
    //fmt.Println(k.Encode())
}
