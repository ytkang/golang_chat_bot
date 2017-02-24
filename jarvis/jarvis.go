package SmartJarvis

import (
	"golang.org/x/net/websocket"
	"github.com/ytkang/golang_chat_bot/network"
	"log"
	"time"
	"strings"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
	"math/rand"
)
type Jarvis struct {
	prevMessageId bson.ObjectId
	mindMap map[string] string
}

func NewJarvis() *Jarvis {
	return &Jarvis{prevMessageId: "", mindMap: make(map[string]string)}
}

func (jarvis *Jarvis) SetPrevMessageId(id bson.ObjectId) {
	jarvis.prevMessageId = id
}

func (jarvis *Jarvis) Answer(ws *websocket.Codec, activeClients map[network.ClientConn]int, message string, db *mgo.Session) {
	time.Sleep(500 * time.Millisecond)
	answer := "hihi~ I'm Jarvis"

	c := db.DB("chat").C("messages")

	if strings.Contains(message, "A:") || strings.Contains(message, "a:") {
		if db == nil {
			return
		}

		message = strings.Replace(message, "A: ", "", 1)
		message = strings.Replace(message, "A:", "", 1)
		message = strings.Replace(message, "a: ", "", 1)
		message = strings.Replace(message, "a:", "", 1)

		log.Println("will learning")
		log.Println(message)
		if jarvis.prevMessageId.Valid() {
			change := mgo.Change{
				Update: bson.M{"$set": bson.M{"text": message}, "$addToSet": bson.M{"answerOf": jarvis.prevMessageId}},
				Upsert: true,
			}

			var obj network.Msg = network.Msg{}
			info, err := c.Find(bson.M{"text": message}).Apply(change, &obj)

			//err = c.Insert(&Msg{Text:clientMessage})
			if err != nil {
				log.Println("[jarvis] find error")
				log.Panic(err)
				answer = "I cannot understand :("
			} else {
				log.Println(info)
				log.Println(obj)
				answer = "I learned about that :)"
			}
		} else {
			answer = "There is no question."
		}
	} else {
		var msgObj network.Msg = network.Msg{}
		var msgObj2 []network.Msg = make([]network.Msg, 0)
		c.Find(bson.M{"text": message}).One(&msgObj)
		log.Println("Found message: ", msgObj)
		c.Find(bson.M{"answerOf": msgObj.ID}).All(&msgObj2)
		if len(msgObj2) == 0 {
			answer = "What is it? please teach me! \"A: your answer\""
		} else {
			rand.Seed(time.Now().Unix())
			i := rand.Intn(len(msgObj2))
			answer = msgObj2[i].Text
		}

	}

	answer = "Jarvis said: "+answer
	log.Println(answer)
	for cs, _ := range activeClients {
		if err := ws.Send(cs.Websocket, answer); err != nil {
			// we could not send the message to a peer

			log.Println("[jarvis] Could not send message to ", cs.ClientIP, err.Error())
		}
	}
}