package poweradapter

import (
    "fmt"
    "math/rand"
)

type Link struct {
    Url string
    TrackingId string
    HitType string
    Title		 //pageview hit type
    Location		 //pageview hit type
    EventCategory string //event hit type
    EventAction string   //event hit type
    EventLabel string    //event hit type
    EventValue int       //event hit type
    CustomPayload string //a user's custom generated hit type
    Payload string `datastore:"-"`
    Hit Hit `datastore:"-"`
}

func (l *Link) Payload() string {
    requiredFields := fmt.Sprintf("v=1&tid=%s&cid=%d",
	l.TrackingId, rand.Intn(1000000))

    // &t= is a required field, but custom hit will supply its own hit type
    if l.HitType == "custom" {
	return requiredFields + l.Hit.String()
    }
    return requiredFields + "&t=" + l.HitType + l.Hit.String()
}

func (l *Link) Load(ps []datastore.Property) err {
    if err := datastore.LoadStruct(l, ps); err != nil {
	return err
    }

    l.Payload := fmt.Sprintf("v=1&tid=%s&cid=%d", l.TrackingId, rand.Intn(1000000))

    switch l.HitType {
    case "pageview":
	l.Payload += fmt.Sprintf("&t=pageview&dt=%s&dl=%s", pageview.Title, pageview.Location)
    case "event":
	el := ""
    	ev := ""
    	if l.EventLabel != "" {
    	    el = "&el=%s" + l.EventLabel
    	}
    	if l.EventValue != 0 {
    	    ev = "&ev=%d" + l.EventValue
    	}
    	l.Payload += fmt.Sprintf("&t=event&ec=%s&ea=%s%s%s", l.EventCategory, l.EventAction, el, ev);
    case "custom":
	l.Payload = l.CustomPayload
    }

    return nil
}

func (l *Link) Save() ([]datastore.Property, err) {
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
	    {
		Name: "Title",
		Value: l.Title,
	    },
	    {
		Name: "Location",
		Value: l.Location,
	    },
	)
    case "event":
	return append(ps,
	    {
		Name: "EventCategory",
		Value: l.EventCategory,
	    },
	    {
		Name: "EventAction",
		Value: l.EventAction,
	    },
	    {
		Name: "EventLabel",
		Value: l.EventLabel,
	    },
	    {
		Name: "EventValue",
		Value: l.EventValue,
	    },
	)
    case "custom":
	return append(ps,
	    {
		Name: "CustomPayload",
		Value: l.CustomPayload,
	    },
	)
    }
}
