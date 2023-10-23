package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	_ "github.com/mcoo/OPQBot/session/provider"
	"github.com/opq-osc/OPQBot/v2"
	"github.com/opq-osc/OPQBot/v2/apiBuilder"
	"github.com/opq-osc/OPQBot/v2/events"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

/*
* @Author <bypanghu> (bypanghu@163.com)
* @Date 2023/7/31 14:48
**/

func IsInGroupS(str int64, groupUids []int64) bool {
	for _, s := range groupUids {
		if s == str {
			return true
		}
	}
	return false
}

func IsAdmins(uid int64, adminUids []int64) bool {
	for _, s := range adminUids {
		if s == uid {
			return true
		}
	}
	return false
}

type setu struct {
	Title string `json:"title"`
	Pic   string `json:"pic"`
}

type moyu struct {
	Success bool   `json:"success"`
	Url     string `json:"url"`
}

func ContainsURL(str string) bool {
	// 定义匹配网址的正则表达式模式
	pattern := `(?i)(https?:\/\/)?([\w-]+\.)?([\w-]+\.[\w-]+)`
	// 编译正则表达式模式
	regex := regexp.MustCompile(pattern)
	// 使用正则表达式进行匹配
	return regex.MatchString(str)
}

func HandleGroupMsg(core *OPQBot.Core) {
	core.On(events.EventNameGroupJoin, func(ctx context.Context, event events.IEvent) {
		groupMsg := event.PraseGroupJoinEvent()
		err := apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUId()).TextMsg("欢迎新人~").Do(ctx)
		if err != nil {
			println(err.Error())
			return
		}
	})

	core.On(events.EventNameGroupExit, func(ctx context.Context, event events.IEvent) {
		groupMsg := event.PraseGroupJoinEvent()
		err := apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUId()).TextMsg("用户退群了~").Do(ctx)
		if err != nil {
			println(err.Error())
			return
		}
	})

	core.On(events.EventNameGroupMsg, func(ctx context.Context, event events.IEvent) {

		if IsInGroupS(event.ParseGroupMsg().GetSenderUin(), banUids) {
			return
		}

		if event.GetMsgType() == events.MsgTypeGroupMsg {
			groupMsg := event.ParseGroupMsg()
			text := groupMsg.ExcludeAtInfo().ParseTextMsg().GetTextContent()

			// 监控的群聊里面禁止所有非管理员发送网址信息
			if IsInGroupS(groupMsg.GetGroupUin(), groupUids) {
				if IsAdmins(groupMsg.GetSenderUin(), adminUids) {
					if ContainsURL(text) {
						wg := sync.WaitGroup{}
						wg.Add(2)
						go func() {
							apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg("发送内容中存在违规网址，系统已撤回，警告一次！").Do(ctx)
							wg.Done()
						}()
						go func() {
							apiBuilder.New(apiUrl, event.GetCurrentQQ()).GroupManager().RevokeMsg().ToGUin(groupMsg.GetGroupUin()).MsgSeq(groupMsg.GetMsgSeq()).MsgRandom(groupMsg.GetMsgRandom()).Do(ctx)
							wg.Done()
						}()
						wg.Wait()
					}
				}

				redbag := groupMsg.GetRedBag()
				if redbag.Wishing != "" {
					err := apiBuilder.New(apiUrl, event.GetCurrentQQ()).GroupManager().GetRedBag().ToGUin(groupMsg.GetGroupUin()).SetRedBagMsg(redbag).Do(ctx)
					if err != nil {
						println(err.Error())
						return
					} else {
						apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg("谢谢老板的红包~~").Do(ctx)
					}
				}

				if text == "可莉日报" {
					apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg("努力获取中，请稍等").DoWithCallBack(ctx, func(iApiBuilder *apiBuilder.Response, err error) {
						localFile, err := downLoadPng("http://api.gauthing.cn/keliribao")
						if err != nil {
							log.Error("下载图片失败", err.Error())
							apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg("我没找到~").Do(ctx)
						}
						// 上传图片到 腾讯服务器
						upload, err := apiBuilder.New(apiUrl, event.GetCurrentQQ()).Upload().SetBase64Buf(localFile).GroupPic().DoUpload(ctx)
						if err != nil {
							log.Error("上传图片失败", err.Error())
						}
						apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).PicMsg(upload).Do(ctx)
					})
				}

				if text == "买家秀" {
					apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg("努力获取中，请稍等").DoWithCallBack(ctx, func(iApiBuilder *apiBuilder.Response, err error) {
						response1, err := iApiBuilder.GetGroupMessageResponse()
						client := &http.Client{}
						baseUrl := "https://api.vvhan.com/api/tao?type=json"
						req, err := http.NewRequest("GET", baseUrl, nil)
						req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
						if err != nil {
							log.Error("拉取图片失败", err.Error())
						}
						response, _ := client.Do(req)
						defer response.Body.Close()
						s, err := ioutil.ReadAll(response.Body)
						var res setu
						err = json.Unmarshal(s, &res)
						if err != nil {
							return
						}
						go apiBuilder.New(apiUrl, event.GetCurrentQQ()).GroupManager().RevokeMsg().ToGUin(groupMsg.GetGroupUin()).MsgSeq(response1.MsgSeq).MsgRandom(response1.MsgTime).Do(ctx)
						// 上传图片到 腾讯服务器
						upload, err := apiBuilder.New(apiUrl, event.GetCurrentQQ()).Upload().SetFileUrlPath(res.Pic).GroupPic().DoUpload(ctx)
						if err != nil {
							log.Error("上传图片失败", err.Error())
						}

						apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).PicMsg(upload).TextMsg(res.Title).DoWithCallBack(ctx, func(iApiBuilder *apiBuilder.Response, err error) {
							response2, err := iApiBuilder.GetGroupMessageResponse()
							if err != nil {
								return
							}
							apiBuilder.New(apiUrl, event.GetCurrentQQ()).GroupManager().RevokeMsg().ToGUin(groupMsg.GetGroupUin()).MsgSeq(groupMsg.GetMsgSeq()).MsgRandom(groupMsg.GetMsgRandom()).Do(ctx)
							time.Sleep(time.Second * 30)
							apiBuilder.New(apiUrl, event.GetCurrentQQ()).GroupManager().RevokeMsg().ToGUin(groupMsg.GetGroupUin()).MsgSeq(response2.MsgSeq).MsgRandom(response2.MsgTime).Do(ctx)
						})
					})

				}

				if text == "美女图" {
					apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg("努力获取中，请稍等").DoWithCallBack(ctx, func(iApiBuilder *apiBuilder.Response, err error) {
						response1, err := iApiBuilder.GetGroupMessageResponse()
						client := &http.Client{}
						baseUrl := "https://api.vvhan.com/api/mobil.girl?type=json"
						req, err := http.NewRequest("GET", baseUrl, nil)
						req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
						if err != nil {
							log.Error("拉取图片失败", err.Error())
						}
						response, _ := client.Do(req)
						defer response.Body.Close()
						s, err := ioutil.ReadAll(response.Body)
						var res struct {
							ImgUrl string `json:"imgurl"`
						}
						err = json.Unmarshal(s, &res)
						if err != nil {
							return
						}
						go apiBuilder.New(apiUrl, event.GetCurrentQQ()).GroupManager().RevokeMsg().ToGUin(groupMsg.GetGroupUin()).MsgSeq(response1.MsgSeq).MsgRandom(response1.MsgTime).Do(ctx)
						localFile, err := downLoadFile(res.ImgUrl)
						if err != nil {
							log.Error("下载图片失败", err.Error())
							apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg("我没找到~").Do(ctx)
						}
						// 上传图片到 腾讯服务器
						upload, err := apiBuilder.New(apiUrl, event.GetCurrentQQ()).Upload().SetBase64Buf(localFile).GroupPic().DoUpload(ctx)
						if err != nil {
							log.Error("上传图片失败", err.Error())
						}
						apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).PicMsg(upload).DoWithCallBack(ctx, func(iApiBuilder *apiBuilder.Response, err error) {
							response2, err := iApiBuilder.GetGroupMessageResponse()
							if err != nil {
								return
							}
							apiBuilder.New(apiUrl, event.GetCurrentQQ()).GroupManager().RevokeMsg().ToGUin(groupMsg.GetGroupUin()).MsgSeq(groupMsg.GetMsgSeq()).MsgRandom(groupMsg.GetMsgRandom()).Do(ctx)
							time.Sleep(time.Second * 30)
							apiBuilder.New(apiUrl, event.GetCurrentQQ()).GroupManager().RevokeMsg().ToGUin(groupMsg.GetGroupUin()).MsgSeq(response2.MsgSeq).MsgRandom(response2.MsgTime).Do(ctx)
						})
					})

				}

				if text == "日历" {
					client := &http.Client{}
					baseUrl := "https://api.vvhan.com/api/moyu?type=json"
					req, err := http.NewRequest("GET", baseUrl, nil)
					req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
					if err != nil {
						log.Error("拉取图片失败", err.Error())
					}
					response, _ := client.Do(req)
					defer response.Body.Close()
					s, err := ioutil.ReadAll(response.Body)
					var res struct {
						Url string `json:"url"`
					}
					err = json.Unmarshal(s, &res)
					if err != nil {
						return
					}
					localFile, err := downLoadFile(res.Url)
					if err != nil {
						log.Error("下载图片失败", err.Error())
						apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg("我没找到~").Do(ctx)
						return
					}

					// 上传图片到 腾讯服务器
					upload, err := apiBuilder.New(apiUrl, event.GetCurrentQQ()).Upload().SetBase64Buf(localFile).GroupPic().DoUpload(ctx)
					if err != nil {
						log.Error("上传图片失败", err.Error())
					}
					apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).PicMsg(upload).Do(ctx)

				}

				if text == "摸鱼日历" {
					client := &http.Client{}
					baseUrl := "https://api.vvhan.com/api/moyu?type=json"
					req, err := http.NewRequest("GET", baseUrl, nil)
					req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
					if err != nil {
						log.Error("拉取图片失败", err.Error())
					}
					response, _ := client.Do(req)
					defer response.Body.Close()
					s, err := ioutil.ReadAll(response.Body)
					var res moyu
					json.Unmarshal(s, &res)

					// 上传图片到 腾讯服务器
					upload, err := apiBuilder.New(apiUrl, event.GetCurrentQQ()).Upload().SetFileUrlPath(res.Url).GroupPic().DoUpload(ctx)
					if err != nil {
						log.Error("上传图片失败", err.Error())
					}
					apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).PicMsg(upload).TextMsg("今日摸鱼日历").Do(ctx)
				}

			}
			// 指令： 禁言@用户 1h (禁言用户和时间中间需要有一个空格)
			if IsAdmins(groupMsg.GetSenderUin(), adminUids) {
				// 禁言用户
				if strings.HasPrefix(text, "禁言") && groupMsg.ContainedAt() {
					atUser := groupMsg.GetAtInfo()
					intTime := 60
					time1 := "1分钟"
					if strings.Contains(text, ":") {
						// 获取禁言时间
						time1 := strings.Split(text, ":")[1]
						// 时间转换为秒
						switch time1 {
						case "1m":
							intTime = 60
						case "1h":
							intTime = 3600
						case "1d":
							intTime = 86400
						case "1w":
							intTime = 604800
						default:
							intTime = 60
						}
					}

					if len(atUser) > 0 {
						group, ctx := errgroup.WithContext(context.Background())
						for _, u := range atUser {
							group.Go(func() error {
								return apiBuilder.New(apiUrl, event.GetCurrentQQ()).GroupManager().ProhibitedUser().ToGUin(groupMsg.GetGroupUin()).ToUid(u.Uid).ShutTime(intTime).Do(ctx)
							})
							group.Go(func() error {
								return apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg(fmt.Sprintf("用户%s已经被成功禁言 %s", u.Nick, time1)).Do(ctx)
							})
						}
						if err := group.Wait(); err != nil {
							apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg("禁言失败，未找到用户").Do(ctx)
						}
					} else {
						apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg("禁言失败，未找到用户").Do(ctx)
					}

				}

				if (strings.HasPrefix(text, "踢") || strings.HasPrefix(text, "移除")) && groupMsg.ContainedAt() {
					user := groupMsg.GetAtInfo()
					if len(user) > 0 {
						group, ctx := errgroup.WithContext(context.Background())
						for _, u := range user {
							group.Go(func() error {
								return apiBuilder.New(apiUrl, event.GetCurrentQQ()).GroupManager().RemoveUser().ToGUin(groupMsg.GetGroupUin()).ToUid(u.Uid).Do(ctx)
							})
							group.Go(func() error {
								return apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg(fmt.Sprintf("用户%s已经被移除本群", u.Nick)).Do(ctx)
							})
						}
						if err := group.Wait(); err != nil {
							apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg("移除失败，未找到用户").Do(ctx)
						}

					} else {
						apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg("移除失败，未找到用户").Do(ctx)
					}
				}

				if strings.HasPrefix(text, "修改昵称") && groupMsg.ContainedAt() {
					user := groupMsg.GetAtInfo()
					newName := strings.Split(text, "为")[1]
					if len(user) > 0 && newName != "" {
						group, ctx := errgroup.WithContext(context.Background())
						for _, u := range user {
							group.Go(func() error {
								return apiBuilder.New(apiUrl, event.GetCurrentQQ()).GroupManager().RenameUserNickName(newName).ToGUin(groupMsg.GetGroupUin()).ToUid(u.Uid).Do(ctx)
							})
							group.Go(func() error {
								return apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg(fmt.Sprintf("用户%s昵称修改完成", u.Nick)).Do(ctx)
							})
						}
						if err := group.Wait(); err != nil {
							apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg(fmt.Sprintf("修改用户昵称失败")).Do(ctx)
						}

					} else {
						apiBuilder.New(apiUrl, event.GetCurrentQQ()).SendMsg().GroupMsg().ToUin(groupMsg.GetGroupUin()).TextMsg("修改失败，未找到用户").Do(ctx)
					}
				}

			}

		}

	})
}
