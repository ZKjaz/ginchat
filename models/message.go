package models

import (
	"context"
	"encoding/json"
	"fmt"
	"main/utils"
	"time"

	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"
)

// 消息
type Message struct {
	gorm.Model
	UserId     int64  //发送者
	TargetId   int64  //接受者  群ID 或者目标用户ID
	Type       int    //发送类型  1私聊  2群聊  3心跳
	Media      int    //消息类型  1文字 2表情包 3语音 4图片 /表情包
	Content    string //消息内容
	CreateTime uint64 //创建时间
	ReadTime   uint64 //读取时间
	Pic        string
	Url        string
	Desc       string
	Amount     int //其他数字统计

}

func (table *Message) TableName() string {
	return "message"
}

// 心跳信息
type Node struct {
	Conn          *websocket.Conn //连接
	Addr          string          //客户端地址
	FirstTime     uint64          //首次连接时间
	HeartbeatTime uint64          //心跳时间
	LoginTime     uint64          //登录时间
	DataQueue     chan []byte     //消息
	GroupSets     set.Interface   //好友 / 群
}

// 映射关系
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁
var rwLocker sync.RWMutex

// 需要：发送者ID,接受者ID,消息类型，发送的内容，发送类型
func Chat(writer http.ResponseWriter, request *http.Request) {
	//1.获取参数 并 检验token等合法性
	//	token:=query.Get("token")

	query := request.URL.Query()
	Id := query.Get("userId")
	userId, _ := strconv.ParseInt(Id, 10, 64)

	//msgType := query.Get("type")
	//targetId := query.Get("targetId")
	//context := query.Get("context")

	//2.获取conn
	isvalida := true
	conn, err := (&websocket.Upgrader{
		//token 校验
		CheckOrigin: func(r *http.Request) bool {
			return isvalida
		},
	}).Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	//3.用户关系
	currentTime := uint64(time.Now().Unix())
	node := &Node{
		Conn:          conn,
		Addr:          conn.RemoteAddr().String(), //客户端地址
		HeartbeatTime: currentTime,                //心跳时间
		LoginTime:     currentTime,                //登录时间
		DataQueue:     make(chan []byte, 50),
		GroupSets:     set.New(set.ThreadSafe),
	}

	//4.userid 跟 node 绑定并加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()

	//5.完成发送的逻辑
	go sendProc(node)
	//6.完成接收的逻辑
	go recvProc(node)

	//sendMsg(userId, []byte("欢迎进入聊天室！！！"))
	//7.加入在线用户到缓存
	SetUserOnlineInfo("online_"+Id, []byte(node.Addr), time.Duration(viper.GetInt("timeout.RedisOnlineTime"))*time.Hour)

}

func sendProc(node *Node) {
	for {

		select {

		case data := <-node.DataQueue:
			fmt.Println("[ws]sendMsg >>>>>> msg:", string(data))
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func recvProc(node *Node) {
	for {

		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		//把数据通过json格式化到msg中
		msg := Message{}
		err = json.Unmarshal(data, &msg)
		if err != nil {
			fmt.Println(err)
		}
		//心跳检测 msg.Media == -1 || msg.Type == 3
		if msg.Type == 3 {
			currentTime := uint64(time.Now().Unix())
			node.Heartbeat(currentTime)
		} else {
			dispatch(data)
			broadMsg(data) //todo 将消息广播到局域网
			fmt.Println("[ws] recvProc <<<<< ", string(data))
		}

	}
}

var udpsendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsendChan <- data
}

func init() {
	go udpsendProc()
	go udpRecvProc()
	fmt.Println("init gorotine")

}

// 完成udp数据发送的协程
func udpsendProc() {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(192, 168, 0, 255),
		Port: viper.GetInt("port.udp"),
	})
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
	}

	for {
		select {
		case data := <-udpsendChan:
			fmt.Println("udpsendProc data:", string(data))

			_, err := conn.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}

}

// 完成udp数据发送的协程
func udpRecvProc() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: viper.GetInt("port.udp"),
	})

	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	for {
		var buf [512]byte

		n, err := conn.Read(buf[:])
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("udpRecvProc data", string(buf[0:n]))
		dispatch(buf[0:n])
	}
}

// 后端调度的逻辑
func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
	}
	switch msg.Type {
	case 1:
		//私信
		fmt.Println("dispatch data:", string(data))
		sendMsg(msg.TargetId, data, 1)
	case 2:
		//群发
		sendGroupMsg(msg.TargetId, data) //发送的群ID ，消息内容
		// case 3:
		//广播
		// 	sendAllMsg()
		// case 4: //心跳
		// node.Heartbeat()
	}
}

func sendMsg(userId int64, msg []byte, t int) {
	//	fmt.Println("sendMsg >>>>> userID: ", userId, "msg :", string(msg))
	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()

	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)

	ctx := context.Background()

	targetIdStr := strconv.Itoa(int(userId))
	userIdStr := strconv.Itoa(int(jsonMsg.UserId))
	jsonMsg.CreateTime = uint64(time.Now().Unix())
	r, err := utils.Red.Get(ctx, "online_"+userIdStr).Result()
	if err != nil {
		fmt.Println(err)
		return
	}
	if r != "" {
		if ok {
			fmt.Println("sendMsg >>> userID: ", userId, "  msg: ", string(msg))
			node.DataQueue <- msg
		}
	}

	//格式化redis消息记录格式，zrevrange msg_1_2_2 0 -1 withscores，查看id为1和2之间缓存的消息，序号由0开始到-1（全部消息）
	var key string
	if userId > jsonMsg.UserId {
		key = "msg_" + userIdStr + "_" + targetIdStr + "_" + strconv.Itoa(t)
	} else {
		key = "msg_" + targetIdStr + "_" + userIdStr + "_" + strconv.Itoa(t)
	}

	res, err := utils.Red.ZRevRange(ctx, key, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
		return
	}
	score := float64(cap(res)) + 1
	ress, e := utils.Red.ZAdd(ctx, key, &redis.Z{score, msg}).Result() //jsonMsg
	//res, e := utils.Red.Do(ctx, "zadd", key, 1, jsonMsg).Result() //备用 后续拓展 记录完整msg
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(ress)

}

// 需要重写此方法才能完整的msg转byte[]
func (msg Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(msg)
}

// 传入目标群ID和msg信息
func sendGroupMsg(targetId int64, msg []byte) {
	fmt.Println("开始群发消息！！！")
	userIds := SearchUserByGroupId(uint(targetId))
	for i := 0; i < len(userIds); i++ {
		// //排除自己
		// if targetId != int64(userIds[i]) {
		// 	sendMsg(int64(userIds[i]), msg)
		// }
		sendMsg(int64(userIds[i]), msg, 2)
	}
}

// 获取缓存里面的消息
func RedisMsg(userIdA int64, userIdB int64, start int64, end int64, t int, isRev bool) []string {
	//rwLocker.RLock()
	//node, ok := clientMap[userIdA]
	//rwLocker.RUnlock()
	//jsonMsg := Message{}
	//json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	userIdStr := strconv.Itoa(int(userIdA))
	targetIdStr := strconv.Itoa(int(userIdB))
	var key string
	if userIdA > userIdB {
		key = "msg_" + targetIdStr + "_" + userIdStr + "_" + strconv.Itoa(t)
	} else {
		key = "msg_" + userIdStr + "_" + targetIdStr + "_" + strconv.Itoa(t)
	}
	//key = "msg_" + userIdStr + "_" + targetIdStr
	//rels, err := utils.Red.ZRevRange(ctx, key, 0, 10).Result()  //根据score倒叙

	var rels []string
	var err error
	if isRev {
		rels, err = utils.Red.ZRange(ctx, key, start, end).Result()
	} else {
		rels, err = utils.Red.ZRevRange(ctx, key, start, end).Result()
	}
	if err != nil {
		fmt.Println(err) //没有找到
	}
	// 发送推送消息
	/**
	// 后台通过websoket 推送消息
	for _, val := range rels {
		fmt.Println("sendMsg >>> userID: ", userIdA, "  msg:", val)
		node.DataQueue <- []byte(val)
	}**/
	return rels
}

// 更新用户心跳
func (node *Node) Heartbeat(currentTime uint64) {
	node.HeartbeatTime = currentTime
	return
}

// 清理超时连接
func CleanConnection(param interface{}) (result bool) {
	result = true
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("cleanConnection err", r)
		}
	}()
	//fmt.Println("定时任务,清理超时连接 ", param)
	//node.IsHeartbeatTimeOut()
	currentTime := uint64(time.Now().Unix())
	for i := range clientMap {
		node := clientMap[i]
		if node.IsHeartbeatTimeOut(currentTime) {
			fmt.Println("心跳超时..... 关闭连接：", node)
			node.Conn.Close()
		}
	}
	return result
}

// 用户心跳是否超时
func (node *Node) IsHeartbeatTimeOut(currentTime uint64) (timeout bool) {
	if node.HeartbeatTime+viper.GetUint64("timeout.HeartbeatMaxTime") <= currentTime {
		fmt.Println("心跳超时。。。自动下线", node)
		timeout = true
	}
	return
}
