package main

import (
	"encoding/json"
	"fmt"
	"fyoukuApi/services/mq"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	beego.LoadAppConfig("ini", "../../conf/app.conf")
	defaultdb := beego.AppConfig.String("defaultdb")
	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", defaultdb, 30, 30)

	mq.ConsumerDlx("fyouku.comment.count", "fyouku.comment.count","fyouku.comment.count.dlx","fyouku.comment.count.dlx",10000, callback)
}

func callback(s string){
	type Data struct {
		VideoId int
		EpisodesId int64
	}
	var data Data
	err := json.Unmarshal([]byte(s), &data)
	if err == nil{
		o := orm.NewOrm()
		//修改视频的总评论数
		o.Raw("update video set comment=comment+1 where id=?", data.VideoId).Exec()
		//修改视频剧集的评论数
		o.Raw("update video_episodes set comment=comment+1 where id=?", data.EpisodesId).Exec()

		//更新redis排行榜-通过MQ来实现
		//创建一个简单模式的MQ
		//把要传递的数据转换为json字符串
		videoObj := map[string]int{
			"VideoId":data.VideoId,
		}
		videoJson,_ := json.Marshal(videoObj)
		mq.Publish("", "fyouku_top", string(videoJson))
	}
	fmt.Printf("msg is :%s\n", s)
}