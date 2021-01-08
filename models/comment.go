package models

import (
	"encoding/json"
	"fmt"
	"fyoukuApi/services/mq"
	"github.com/astaxie/beego/orm"
	"time"
)

type Comment struct {
	Id          int
	Content     string
	AddTime     int64
	UserId      int
	Stamp       int
	Status      int
	PraiseCount int
	EpisodesId  int
	VideoId     int
}

func init() {
	orm.RegisterModel(new(Comment))
}

func SaveComment(content string, uid int, episodesId int, videoId int) error {
	o := orm.NewOrm()
	var comment Comment
	comment.Content = content
	comment.UserId = uid
	comment.EpisodesId = episodesId
	comment.VideoId = videoId
	comment.Stamp = 0
	comment.Status = 1
	comment.AddTime = time.Now().Unix()

	_, err := o.Insert(&comment)
	if err == nil {
		//修改视频的总评论数
		o.Raw("update video set comment=comment+1 where id=?", videoId).Exec()
		//修改视频剧集的评论数
		o.Raw("update video_episodes set comment=comment+1 where id=?", episodesId).Exec()

		//更新redis排行榜-通过MQ来实现
		//创建一个简单模式的MQ
		//把要传递的数据转换为json字符串
		videoObj := map[string]int{
			"VideoId":videoId,
		}
		videoJson,_ := json.Marshal(videoObj)
		mq.Publish("", "fyouku_top", string(videoJson))

		//延迟队列增加评论数
		videoCountObj := map[string]int{
			"VideoId":videoId,
			"EpisodesId":episodesId,
		}
		videoCountJson,_ := json.Marshal(videoCountObj)
		err = mq.PublishDlx("fyouku.comment.count", string(videoCountJson))
		fmt.Println(err)
	}
	return err
}

func GetCommentList(episodesId int, offset int, limit int) (int64, []Comment, error) {
	o := orm.NewOrm()
	var comments []Comment
	num, _ := o.Raw("select id from comment where status=1 and episodes_id=?", episodesId).QueryRows(&comments)
	_, err := o.Raw("select id,content,add_time,user_id,stamp,praise_count,episodes_id from comment where status=1 and episodes_id=? order by add_time desc limit ?,?", episodesId, offset, limit).QueryRows(&comments)
	return num, comments, err
}
