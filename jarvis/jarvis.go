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

	c := db.DB("ytchat").C("messages")

	if strings.Contains(message, "ㄷ:") {
		if db == nil {
			return
		}

		message = strings.Replace(message, "ㄷ: ", "", 1)
		message = strings.Replace(message, "ㄷ:", "", 1)

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
				answer = "으악 몰람.. 모르겠어 :("
			} else {
				log.Println(info)
				log.Println(obj)
				answer = "오호? 땡큐베리마취! 잘 배웠어욧! :)"
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
			answer = "오잉? 이게 뭐지!? 답변좀 굽신굽신~! \"ㄷ:\"를 사용하면 됩니다! 예) ㄷ: ㅋㅋㅋ 이거야~"
		} else {
			rand.Seed(time.Now().Unix())
			i := rand.Intn(len(msgObj2))
			answer = msgObj2[i].Text
		}

	}

	answer = "Jarvis Said: "+answer
	log.Println(answer)
	for cs, _ := range activeClients {
		if err := ws.Send(cs.Websocket, answer); err != nil {
			// we could not send the message to a peer
			log.Println("[jarvis] Could not send message to ", cs.ClientIP, err.Error())
		}
	}
}