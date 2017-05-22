package main

import (
	"fmt"

	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
	"gopkg.in/kataras/iris.v6/adaptors/websocket"
	"github.com/golang/protobuf/proto"
	"github.com/xmxiaoq/ffws/pb"
	"log"
	"encoding/binary"
)

/* Native messages no need to import the iris-ws.js to the ./templates.client.html
Use of: OnMessage and EmitMessage.


NOTICE: IF YOU HAVE RAN THE PREVIOUS EXAMPLES YOU HAVE TO CLEAR YOUR BROWSER's CACHE
BECAUSE chat.js is different than the CACHED. OTHERWISE YOU WILL GET Ws is undefined from the browser's console, becuase it will use the cached.
*/

type clientPage struct {
	Title string
	Host  string
}

func main() {
	app := iris.New()
	app.Adapt(iris.DevLogger()) // enable all (error) logs
	app.Adapt(httprouter.New()) // select the httprouter as the servemux
	//app.Adapt(view.HTML("./templates", ".html")) // select the html engine to serve templates

	ws := websocket.New(websocket.Config{
		// the path which the websocket client should listen/registered to,
		Endpoint: "/ws",
		// to enable binary messages (useful for protobuf):
		BinaryMessages: true,
	})

	app.Adapt(ws) // adapt the websocket server, you can adapt more than one with different Endpoint

	//app.StaticWeb("/js", "./static/js") // serve our custom javascript code

	//app.Get("/", func(ctx *iris.Context) {
	//	ctx.Render("client.html", clientPage{"Client Page", ctx.Host()})
	//})

	ws.OnConnection(func(c websocket.Connection) {
		fmt.Println("连接成功:", c.ID())

		c.OnMessage(func(data []byte) {
			//message := string(data)
			//c.To(websocket.Broadcast).EmitMessage([]byte("Message from: " + c.ID() + "-> " + message)) // broadcast to all clients except this
			//c.EmitMessage([]byte("Me: " + message))                                                    // writes to itself
			//c.EmitMessage(data)

			b := data
			sz := binary.LittleEndian.Uint32(b[0:4])
			protoID := pb.C2S_ProtoId(binary.LittleEndian.Uint32(b[4:8]))
			log.Println(sz, protoID, len(b))
			if protoID == pb.C2S_ProtoId_cLogin {
				var req pb.LoginReq
				err := proto.Unmarshal(b[8:], &req)
				if err != nil {
					log.Println(err)
				} else {
					log.Println("客户端发来数据:", req)
					var resp pb.LoginRsp
					resp.Ret = req.Uid
					var uid int32 = 101
					resp.Info = &pb.UserInfo{Uid:&uid}
					bt, err := proto.Marshal(&resp)
					if err != nil {
						log.Println(err)
					} else {
						btSend := make([]byte, 8, 8+len(bt))
						binary.LittleEndian.PutUint32(btSend, uint32(4+len(bt)))
						binary.LittleEndian.PutUint32(btSend[4:], uint32(pb.S2C_ProtoId_sLogin))
						btSend = append(btSend, bt...)
						c.EmitMessage(btSend)
						fmt.Println(btSend)
					}
				}
			}
		})

		c.OnDisconnect(func() {
			fmt.Printf("\nConnection with ID: %s has been disconnected!\n", c.ID())
		})

	})

	app.Listen(":8080")

}
