package main
import (
	"time"
	"log"
	"fmt"
	"github.com/kataras/iris/v12"
	lp "asd/longpoll"
)
type eventMsg struct {
	Id int64 `json:"id,omitempty"`
	To string `json:"to,omitempty"`
	From string `json:"from,omitempty"`
	Text string `json:"text,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	CreatedTime int64 `json:"createdTime,omitempty"`
	ContentMetadata map[string]string `json:"contentMetadata,omitempty"`
	Chunks [][]byte `json:"chunks"`
	ContentPreview []byte `json:"preview"`
}
type eventType int64
func eventTypeFromString(e string) (eventType, bool) {
	if e == "END_OF_OPERATION"	{ return 0, true }
	if e == "ONLINE"			{ return 1, true }
	if e == "OFFLINE"			{ return 2, true }
	if e == "ACCOUNT_CREATE"	{ return 3, true }
	if e == "ACCOUNT_SUSPEND"	{ return 4, true }
	if e == "ACCOUNT_DELETE"	{ return 5, true }
	if e == "ACCOUNT_MODIFIED"	{ return 6, true }

	if e == "SETTING_UPDATE"	{ return 7, true }
	if e == "PROFILE_UPDATE"	{ return 8, true }
	if e == "GROUP_CREATE"		{ return 9, true }
	if e == "GROUP_UPDATE"		{ return 10, true }
	if e == "GROUP_LEAVE"		{ return 11, true }
	if e == "GROUP_INVITE"		{ return 12, true }
	if e == "GROUP_INVITED"		{ return 13, true }
	if e == "GROUP_KICK"		{ return 14, true }
	if e == "GROUP_KICKED"		{ return 15, true }

	if e == "SEND_MESSAGE"		{ return 25, true }
	if e == "RECV_MESSAGE"		{ return 26, true }
	if e == "UNSEND_MESSAGE"	{ return 27, true }
	if e == "EDIT_MESSAGE"		{ return 28, true }

	if e == "NEED_LOGIN"		{ return 88, true }
	if e == "LOGOUT"			{ return 89, true }

	return -1, false
}
type eventData struct {
	Offset int64 `json:"offset"`
	CreatedTime int64 `json:"createdTime,omitempty"`
	Type eventType `json:"type,omitempty"`
	Param1 string `json:"param1,omitempty"`
	Param2 string `json:"param2,omitempty"`
	Param3 string `json:"param3,omitempty"`
	Message *eventMsg `json:"message,omitempty"`
}
type APIResponse struct {
	Code int `json:"code,omitempty"`
	Reason string `json:"reason,omitempty"`
	Events []eventData `json:"events,omitempty"`
	Offset int64 `json:"offset,omitempty"`
	NextOffset int64 `json:"nextOffset"`
}

func EpochMsNow() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
func MaxVal(a, b int64) int64 {
	if a > b { return a }
	return b
}
func isset(total, index int) bool {
	return total > index 
}
func newEvent(user string, event eventData) eventData {
	event.Offset = int64(len(userEventData[user])) 
	event.CreatedTime = EpochMsNow()
	userEventData[user] = append(userEventData[user], event)
	return event
}

// [user][msgId]
var userMessage = map[string]map[string]eventMsg{}
// [user][0]
var userEventData = map[string][]eventData{}

func main() {
	lp.DEBUG = true
	lpMgr, err := lp.NewManager(lp.Options{
		// IDK WHY THIS BUGGY
		//MaxConnection: 5, // lp.UNLIMITED_CONN,
		//TimeoutSeconds: 20, //lp.UNSET,
		MaxConnection: lp.UNLIMITED_CONN,
		TimeoutSeconds: lp.UNSET,
	})
	if err != nil {
		panic(err)
	}
	app := iris.New()

	app.Get("/", func(ctx iris.Context){
		ctx.Writef("<html><body><h3>Select user:</h3><a href=\"anysz\">anysz</a><br /><a href=\"lore\">lore</a><br /><a href=\"ipsum\">ipsum</a></body></html>")
	})

	app.Get("/{username:string}", func(ctx iris.Context){
		// act like create an account and/or login
		name := ctx.Params().GetString("username")
		if _, exist := userEventData[name]; !exist {
			ed := new(eventData)
			ed.Offset = 1
			ed.CreatedTime = EpochMsNow()
			ed.Type, _ = eventTypeFromString("ACCOUNT_CREATE")
			userEventData[name] = []eventData{*ed}
			userMessage[name] = map[string]eventMsg{}
		}
		ctx.Writef("%s", fmt.Sprintf(`<!DOCTYPE html>
<html>
<!-- this is just demo app -->
<head><title>%s</title></head>
<body>
<h1>Welcome, %s</h1>
	<div>Event test: <input type="text" name="to_user" placeholder="lore" /> <input type="text" name="event_name" placeholder="PROFILE_UPDATE" /> <input type="button" value="Create!" onclick="eventSender()" /></div>
	<div name="event_container"></div>
	<script>
	function setContainer(content) {
		document.getElementsByName("event_container")[0].innerHTML = content
	}
	function addToContainer(content) {
		document.getElementsByName("event_container")[0].innerHTML += content
	}
	let username = "%s"
	addToContainer("<b>Client Event started</b><br />")
	let glbError = (url) => function() { console.log(url + ": request error") }
	function get(url, onload) {
		let xhr = new XMLHttpRequest()
		xhr.open("GET", url)
		xhr.onload = function(){
			onload(xhr)
		}
		xhr.onerror = glbError(url)
		xhr.send();
	}

	function eventSender() {
		let username = document.getElementsByName("to_user")[0].value
		let eventName = document.getElementsByName("event_name")[0].value
		get("/new_event/"+username+"/"+eventName, function(xhr){
			addToContainer("<b>Sending Event ["+eventName+"] to ["+username+"]: " + xhr.response + "</b><br />")
		})
	}

	// getting last event offset
	get("/event/"+username+"/getLastOffset", function(xhr){
		addToContainer("<b>Latest event offset:"+(xhr.response)+"</b><br />")
	})
	function fetchEventLongPolling(offset){
		get("/event/"+username+"/"+offset+"/1", function(xhr){			
			addToContainer("<b>Fetching offset "+offset+" with limit 1:</b><br />")
			try{
				r = JSON.parse(xhr.response)
				for(i=0; i < r.events.length; i++){
					addToContainer("<b>Event:&nbsp;</b>" + JSON.stringify(r.events[i]) + "<br />")
				}
				addToContainer("<b>Next event offset revision: </b>" + r.nextOffset + "<br />")
				offset = r.nextOffset
			}catch(e){
				addToContainer("<b>Error: "+xhr.response+"</b><br />")
			}
			fetchEventLongPolling(offset)
		})
	}
	addToContainer("<b>App ok. fetching now.</b><br />")
	fetchEventLongPolling(0)
	</script>
</body>
</html>`, name, name, name))
	})

	app.Get("/new_event/{username:string}/{eventTypeString:string}", func(ctx iris.Context){
		// act like push event data
		name := ctx.Params().GetString("username")
		eventTypeString := ctx.Params().GetString("eventTypeString")
		etype, isExist := eventTypeFromString(eventTypeString)
		if !isExist {
			ctx.Writef("%s", "404: event type name not found")
			return
		}
		ed := new(eventData)
		ed.Type = etype
		isError := lpMgr.Publish(name, newEvent(name, *ed))
		ctx.JSON(struct{
			error bool `json:"error,omitempty"`
		}{isError})
	})

	app.Get("/event/{eventName:string}/getLastOffset", func(ctx iris.Context){
		name := ctx.Params().GetString("eventName")
		rs := new(APIResponse)
		events, exist := userEventData[name]
		if !exist {
			ed := new(eventData)
			ed.Offset = -1
			ed.CreatedTime = EpochMsNow()
			ed.Type, _ = eventTypeFromString("NEED_LOGIN")
			rs.Code = 401
			rs.Events = append(rs.Events, *ed)
			ctx.JSON(rs)
			return
		}
		rs.Code = 200
		rs.Offset = int64(len(events))
		ctx.JSON(rs)
	})
	app.Get("/event/{eventName:string}/{rev:int64}/{limit:int64}", func(ctx iris.Context){
		// act like getting user event based on event Offset and Limit
		name := ctx.Params().GetString("eventName")
		rev, _ := ctx.Params().GetInt64("rev")
		limit, _ := ctx.Params().GetInt64("limit")
		if rev < 0 { rev = 0 }
		if limit < 0 { limit = 0 }
		if limit > 50 { limit = 50 }
		rs := new(APIResponse)
		rs.Code = 200
		_, exist := userEventData[name]
		if !exist {
			// just act like authentication session
			ed := new(eventData)
			ed.Offset = -1
			ed.CreatedTime = EpochMsNow()
			ed.Type, _ = eventTypeFromString("NEED_LOGIN")
			rs.Code = 401
			rs.Events = append(rs.Events, *ed)
			ctx.JSON(rs)
			return
		}
		total := len(userEventData[name])
		lastOffset := int64(0)
		if !isset(total, int(rev)) {
			// data is not exist, waiting for new data
			data, isInvalid	:= lpMgr.Subscribe(name)
			if isInvalid {
				// Timeout or too much polling client
				ctx.StatusCode(iris.StatusGone)
				log.Println(name, "timeout, client need to resend")
				ctx.JSON(APIResponse{
					Code: 410,
					Reason: "longpolling timedout, resend the request.",
					Events: []eventData{},
					NextOffset: rev,
				})
				return
			}
			lastOffset = MaxVal(data.(eventData).Offset, lastOffset)
			log.Println("event comming", data.(eventData))
			rs.Events = append(rs.Events, data.(eventData))
		}else{
			// first data exist, continue to next event
			for i := int64(0); i < limit; i++ {
				if isset(total, int(rev+i)) {
					// if the data already exist
					event := userEventData[name][rev+i]
					lastOffset = MaxVal(event.Offset, lastOffset)
					rs.Events = append(rs.Events, event)
					log.Println("found", i, event)
				}
			}
		}
		if limit != 0 && !isset(total, int(rev+limit)) {
			// to inform to the client is next data exist or no
			// if there's end_of_operation, next revision request will wait until new data arrived
			endEvent, _ := eventTypeFromString("END_OF_OPERATION")
			rs.Events = append(rs.Events, eventData{
				Offset: -1,
				Type: endEvent,
			})
		}

		// next token
		rs.NextOffset = lastOffset+1
		ctx.JSON(rs)
	})

	app.Run(iris.Addr("127.0.0.1:8823"),
   		iris.WithoutServerError(iris.ErrServerClosed),
    	iris.WithOptimizations,
  	)
}
