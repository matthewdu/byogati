package poweradapter

import (
    "fmt"
    "math/rand"

    "google.golang.org/appengine/datastore"
)

type Link struct {
    Url string `json:"url" binding:"required"`
    TrackingId string `json:"trackingId" binding:"required"`
    HitType string `json:"hitType" binding:"required"`
    Title string `json:"title"`        //pageview hit type
    Location string `json:"location"`     //pageview hit type
    EventCategory string `json:"eventCategory"`//event hit type
    EventAction string `json:"eventAction"`  //event hit type
    EventLabel string `json:"eventLabel"`   //event hit type
    EventValue int `json:"eventValue"`      //event hit type
    CustomPayload string `json:"CustomPayload"`//a user's custom generated hit type
    Payload string `datastore:"-"` //auto generated from PropertyLoadSaver.Load
}

func (l *Link) Load(ps []datastore.Property) error {
    if err := datastore.LoadStruct(l, ps); err != nil {
	return err
    }

    l.Payload = fmt.Sprintf("v=1&tid=%s&cid=%d", l.TrackingId, rand.Intn(1000000))

    switch l.HitType {
    case "pageview":
        l.Payload += fmt.Sprintf("&t=pageview&dt=%s&dl=%s", l.Title, l.Location)
    case "event":
        el := ""
    	ev := ""
    	if l.EventLabel != "" {
    	    el = fmt.Sprintf("&el=%s", l.EventLabel)
    	}
    	if l.EventValue != 0 {
    	    ev = fmt.Sprintf("&ev=%d", l.EventValue)
    	}
    	l.Payload += fmt.Sprintf("&t=event&ec=%s&ea=%s%s%s", l.EventCategory, l.EventAction, el, ev);
    case "custom":
        l.Payload = l.CustomPayload
    }

    return nil
}

func (l *Link) Save() ([]datastore.Property, error) {
    ps := []datastore.Property{
	{
	    Name: "Url",
	    Value: l.Url,
	},
	{
	    Name: "TrackingId",
	    Value: l.TrackingId,
	},
	{
	    Name: "HitType",
	    Value: l.HitType,
	},
    }

    switch l.HitType {
    case "pageview":
	return append(ps,
	    datastore.Property{
		Name: "Title",
		Value: l.Title,
	    },
	    datastore.Property{
		Name: "Location",
		Value: l.Location,
	    },
	), nil
    case "event":
	return append(ps,
	    datastore.Property{
		Name: "EventCategory",
		Value: l.EventCategory,
	    },
	    datastore.Property{
		Name: "EventAction",
		Value: l.EventAction,
	    },
	    datastore.Property{
		Name: "EventLabel",
		Value: l.EventLabel,
	    },
	    datastore.Property{
		Name: "EventValue",
		Value: int64(l.EventValue),
	    },
	), nil
    case "custom":
	return append(ps,
	    datastore.Property{
		Name: "CustomPayload",
		Value: l.CustomPayload,
	    },
	), nil
    }

    panic("No HitType")
    return ps, nil
}
