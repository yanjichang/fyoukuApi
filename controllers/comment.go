package controllers

import (
	"fyoukuApi/models"
	"github.com/astaxie/beego"
)

type CommentController struct {
	beego.Controller
}

type CommentInfo struct {
	Id           int             `json:"id"`
	Content      string          `json:"content"`
	AddTime      int64           `json:"addTime"`
	AddTimeTitle string          `json:"addTimeTitle"`
	UserId       int             `json:"userId"`
	Stamp        int             `json:"stamp"`
	PraiseCount  int             `json:"praiseCount"`
	UserInfo     models.UserInfo `json:"userinfo"`
	EpisodesId   int             `json:"episodesId"`
}

//发表评论
// @router /comment/save [*]
func (this *CommentController) Save() {
	content := this.GetString("content")
	uid, _ := this.GetInt("uid")
	episodesId, _ := this.GetInt("episodesId")
	videoId, _ := this.GetInt("videoId")

	if content == "" {
		this.Data["json"] = ReturnError(4001, "内容不能为空")
		this.ServeJSON()
	}
	if uid == 0 {
		this.Data["json"] = ReturnError(4002, "请先登录")
		this.ServeJSON()
	}
	if episodesId == 0 {
		this.Data["json"] = ReturnError(4003, "必须指定评论剧集ID")
		this.ServeJSON()
	}
	if videoId == 0 {
		this.Data["json"] = ReturnError(4005, "必须指定视频ID")
		this.ServeJSON()
	}

	err := models.SaveComment(content, uid, episodesId, videoId)
	if err == nil {
		this.Data["json"] = ReturnSuccess(0, "success", "", 1)
		this.ServeJSON()
	} else {
		this.Data["json"] = ReturnError(5000, err)
		this.ServeJSON()
	}

}

//获取评论列表 goroutine
// @router /comment/list [*]
func (this *CommentController) List() {
	episodesId, _ := this.GetInt("episodesId")
	limit, _ := this.GetInt("limit")
	offset, _ := this.GetInt("offset")

	if episodesId == 0 {
		this.Data["json"] = ReturnError(4001, "")
		this.ServeJSON()
	}
	if limit == 0 {
		limit = 12
	}
	num, comments, err := models.GetCommentList(episodesId, offset, limit)
	if err == nil {
		var data []CommentInfo
		var commentInfo CommentInfo

		//获取uid channel
		uidChan := make(chan int,12)
		closeChan := make(chan bool,5)
		resChan := make(chan models.UserInfo, 12)

		//把获取的uid 放到channel中
		go func(){
			for _,v := range comments{
				uidChan <- v.UserId
			}
			close(uidChan)
		}()

		//处理uidChannel中的消息
		for i:=0; i<5; i++{
			go chanGetUserInfo(uidChan, resChan, closeChan)
		}

		//判断是否执行完成 信息聚合
		go func(){
			for i:=0; i<5; i++{
				<-closeChan
			}
			close(resChan)
			close(closeChan)
		}()
		userInfoMap := make(map[int]models.UserInfo)
		for r:= range resChan{
			userInfoMap[r.Id] = r
		}
		for _, v := range comments {
			commentInfo.Id = v.Id
			commentInfo.Content = v.Content
			commentInfo.AddTime = v.AddTime
			commentInfo.AddTimeTitle = DateFormat(v.AddTime)
			commentInfo.UserId = v.UserId
			commentInfo.Stamp = v.Stamp
			commentInfo.PraiseCount = v.PraiseCount
			commentInfo.EpisodesId = v.EpisodesId
			commentInfo.UserInfo, _ = userInfoMap[v.UserId]
			data = append(data, commentInfo)
		}
		this.Data["json"] = ReturnSuccess(0, "success", data, num)
		this.ServeJSON()
	} else {
		this.Data["json"] = ReturnError(4004, "没有相关内容")
		this.ServeJSON()
	}
}

func chanGetUserInfo(uidChan chan int, resChan chan models.UserInfo, closeChan chan bool){
	for uid := range uidChan{
		res,err := models.GetUserInfo(uid)
		if err == nil{
			resChan <- res
		}
	}
	closeChan <- true
}


//func (this *CommentController) List() {
//	episodesId, _ := this.GetInt("episodesId")
//	limit, _ := this.GetInt("limit")
//	offset, _ := this.GetInt("offset")
//
//	if episodesId == 0 {
//		this.Data["json"] = ReturnError(4001, "")
//		this.ServeJSON()
//	}
//	if limit == 0 {
//		limit = 12
//	}
//	num, comments, err := models.GetCommentList(episodesId, offset, limit)
//	if err == nil {
//		var data []CommentInfo
//		var commentInfo CommentInfo
//		for _, v := range comments {
//			commentInfo.Id = v.Id
//			commentInfo.Content = v.Content
//			commentInfo.AddTime = v.AddTime
//			commentInfo.AddTimeTitle = DateFormat(v.AddTime)
//			commentInfo.UserId = v.UserId
//			commentInfo.Stamp = v.Stamp
//			commentInfo.PraiseCount = v.PraiseCount
//			commentInfo.EpisodesId = v.EpisodesId
//			commentInfo.UserInfo, _ = models.RedisGetUserInfo(v.UserId)
//			data = append(data, commentInfo)
//		}
//		this.Data["json"] = ReturnSuccess(0, "success", data, num)
//		this.ServeJSON()
//	} else {
//		this.Data["json"] = ReturnError(4004, "没有相关内容")
//		this.ServeJSON()
//	}
//}