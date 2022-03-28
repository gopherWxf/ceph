package locate

import (
	"ceph/chapter5/lib/RabbitMQ"
	"ceph/chapter5/lib/rs"
	"ceph/chapter5/lib/types"
	"encoding/json"
	"errors"
	"os"
	"time"
)

//判断收到的数据分片是否大于等于要求，true则说明对象存在
func Exist(hash string) bool {
	locateinfo, _ := Locate(hash)
	return len(locateinfo) >= rs.DATA_SHARDS
}

//查找哪几台数据节点存了该object的数据分片
//locateInfo map[int]string key是分片的id，val是该分片的数据节点的地址
func Locate(hash string) (locateInfo map[int]string, err error) {
	//创建一个rabbitmq结构体的实例
	r := RabbitMQ.NewRabbitMQ(os.Getenv("RABBITMQ_SERVER"))
	defer r.Close()
	//将object发布到dataServers这个exchange里面去，供别人接收
	//所有绑定dataServers这个exchange的队列都会收到这个消息
	r.Publish("dataServers", hash)
	//获取消费队列的channel
	ch := r.Consume()
	//等待一秒钟
	go func() {
		time.Sleep(1 * time.Second)
		r.Close()
	}()
	locateInfo = make(map[int]string)
	for i := 0; i < rs.ALL_SHARDS; i++ {
		msg := <-ch
		if len(string(msg.Body)) == 0 {
			return nil, errors.New("msg.Body an error occurred, no message")
		}
		var info types.LocateMessage
		json.Unmarshal(msg.Body, &info)
		locateInfo[info.Id] = info.Addr
	}
	return
}