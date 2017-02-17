package network

import (
	"golang.org/x/net/websocket"
	"gopkg.in/mgo.v2/bson"
)

type (
	ClientConn struct {
		Websocket *websocket.Conn
		ClientIP  string
	}

	Msg struct {
		ID		bson.ObjectId `bson:"_id,omitempty"`
		Text		string	`bson:"text"`
		AnswerOf	[]bson.ObjectId `bson:"answerOf,omitempty"`
	}
)
