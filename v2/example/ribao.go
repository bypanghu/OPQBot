package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/jasonlvhit/gocron"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/*
* @Author <bypanghu> (bypanghu@163.com)
* @Date 2023/8/3 16:33
**/

//go:embed  html/*
var htmlFIle embed.FS

type BilibiliHotWord struct {
	ExpStr string `json:"exp_str"`
	Code   int    `json:"code"`
	Cost   struct {
		ParamsCheck           string `json:"params_check"`
		HotwordRequest        string `json:"hotword_request"`
		HotwordRequestFormat  string `json:"hotword_request_format"`
		HotwordResponseFormat string `json:"hotword_response_format"`
		DeserializeResponse   string `json:"deserialize_response"`
		Total                 string `json:"total"`
		MainHandler           string `json:"main_handler"`
	} `json:"cost"`
	Seid      string `json:"seid"`
	Timestamp int    `json:"timestamp"`
	Message   string `json:"message"`
	List      []struct {
		Status     string  `json:"status"`
		CallReason int     `json:"call_reason"`
		HeatLayer  string  `json:"heat_layer"`
		HotId      int     `json:"hot_id"`
		Keyword    string  `json:"keyword"`
		ResourceId int     `json:"resource_id"`
		GotoType   int     `json:"goto_type"`
		ShowName   string  `json:"show_name"`
		Pos        int     `json:"pos"`
		WordType   int     `json:"word_type"`
		Id         int     `json:"id"`
		Score      float64 `json:"score"`
		GotoValue  string  `json:"goto_value"`
		StatDatas  struct {
			Reply1H            string `json:"reply_1h"`
			ShareRt            string `json:"share_rt"`
			PlayTotalRank1HDiv string `json:"play_total_rank_1h_div"`
			Mtime              string `json:"mtime"`
			PlayRt             string `json:"play_rt"`
			Share1H            string `json:"share_1h"`
			Danmu1H            string `json:"danmu_1h"`
			PosStart           string `json:"pos_start"`
			Category           string `json:"category"`
			RelatedResource    string `json:"related_resource"`
			PosType            string `json:"pos_type"`
			Source             string `json:"source"`
			Etime              string `json:"etime"`
			DanmuRt            string `json:"danmu_rt"`
			LikesRt            string `json:"likes_rt"`
			IsCommercial       string `json:"is_commercial"`
			ScoreExp           string `json:"score_exp"`
			Likes1H            string `json:"likes_1h"`
			Stime              string `json:"stime"`
			Rankscore          string `json:"rankscore"`
			PosEnd             string `json:"pos_end"`
			ReplyRt            string `json:"reply_rt"`
			Play1H             string `json:"play_1h"`
			CardType           string `json:"card_type"`
		} `json:"stat_datas"`
		LiveId   []interface{} `json:"live_id"`
		NameType string        `json:"name_type"`
		Icon     string        `json:"icon"`
	} `json:"list"`
	HotwordEggInfo string `json:"hotword_egg_info"`
}

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel struct {
		Title string `xml:"title"`
		Item  []struct {
			Title   string `xml:"title"`
			PubDate string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

type RssJson struct {
	Title   string `json:"title"`
	PubDate string `json:"pubDate"`
}

type Fan struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Result  struct {
		List []struct {
			Badge string `json:"badge"`
			Cover string `json:"cover"`
			Title string `json:"title"`
			Url   string `json:"url"`
		} `json:"list"`
		Note string `json:"note"`
	} `json:"result"`
}

func FetchBilibiliFanList() (*Fan, error) {
	client := &http.Client{}
	baseUrl := "https://api.bilibili.com/pgc/web/rank/list?day=3&season_type=1"
	req, err := http.NewRequest("GET", baseUrl, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
	if err != nil {
		log.Fatalf("拉取图片失败: %s", err.Error())
		return nil, err
	}
	response, _ := client.Do(req)
	defer response.Body.Close()
	s, err := ioutil.ReadAll(response.Body)
	var res Fan
	json.Unmarshal(s, &res)
	return &res, nil
}

func FetchBilibiliHotWord() (*BilibiliHotWord, error) {
	client := &http.Client{}
	baseUrl := "https://s.search.bilibili.com/main/hotword"
	req, err := http.NewRequest("GET", baseUrl, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
	if err != nil {
		log.Fatalf("拉取图片失败: %s", err.Error())
		return nil, err
	}
	response, _ := client.Do(req)
	defer response.Body.Close()
	s, err := ioutil.ReadAll(response.Body)
	var res BilibiliHotWord
	json.Unmarshal(s, &res)
	return &res, nil
}

func FetchTeck() ([]byte, error) {
	client := &http.Client{}
	baseUrl := "https://www.ithome.com/rss/"
	req, err := http.NewRequest("GET", baseUrl, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
	if err != nil {
		log.Fatalf("拉取图片失败: %s", err.Error())
		return nil, err
	}
	response, _ := client.Do(req)
	defer response.Body.Close()
	s, err := ioutil.ReadAll(response.Body)
	var res Rss
	xml.Unmarshal(s, &res)
	var RssJsons []RssJson

	if res.Channel.Item != nil {
		for _, key := range res.Channel.Item {
			RssJsons = append(RssJsons, RssJson{Title: key.Title, PubDate: key.PubDate})
		}
	}
	// 将结构体转换为JSON格式
	data, err := json.Marshal(RssJsons)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func handleWebServe(openApi bool) {
	// 定义文件服务器的根目录
	fs := http.FileServer(http.FS(htmlFIle))
	// 注册处理器来处理根路径的请求
	http.Handle("/", fs)
	http.Handle("/bilibiliHotWord", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if openApi {
			// 设置跨域标头
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		word, err := FetchBilibiliHotWord()
		// 将结构体转换为JSON格式
		data, err := json.Marshal(word)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 设置HTTP头为application/json
		w.Header().Set("Content-Type", "application/json")
		// 写入JSON响应
		w.Write(data)
	}))

	http.Handle("/bilibiliFanList", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if openApi {
			// 设置跨域标头
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		word, err := FetchBilibiliFanList()
		// 将结构体转换为JSON格式
		data, err := json.Marshal(word)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 设置HTTP头为application/json
		w.Header().Set("Content-Type", "application/json")
		// 写入JSON响应
		w.Write(data)
	}))

	http.Handle("/getTeckList", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if openApi {
			// 设置跨域标头
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		word, err := FetchTeck()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 设置HTTP头为application/json
		w.Header().Set("Content-Type", "application/json")
		// 写入JSON响应
		w.Write(word)
	}))

	if openApi {
		http.Handle("/keliribao", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 设置跨域标头
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			today := time.Now().Format("2006-01-02")
			fileName := fmt.Sprintf("screenLock-%s.png", today)
			// 判断是否有文件，如果有就 直接返回数据
			if _, err := os.Stat(fileName); err == nil {
				fmt.Println("fileName has exist , return it")
				// 直接返回文件流
				http.ServeFile(w, r, fileName)
				return
			}
			screenLock(fileName)
			// 直接返回文件流
			http.ServeFile(w, r, fileName)
			return
		}))
	}

	// 启动HTTP服务器并监听在指定的端口上
	log.Println("Server started on http://localhost:10088")
	http.ListenAndServe(":10088", nil)
}

func screenLock(fileName string) {
	// 创建一个上下文和取消函数
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	// 设置浏览器选项
	options := []chromedp.ExecAllocatorOption{
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.NoFirstRun,
	}
	// 创建一个浏览器执行器
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, options...)
	defer cancel()

	// 创建一个新的浏览器上下文
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()
	// 加载网页
	var buf []byte
	if err := chromedp.Run(ctx,
		chromedp.EmulateViewport(800, 1850),
		chromedp.Navigate("http://localhost:10088/html/ribao.html"),
		// 睡眠2秒钟，保证页面图片渲染完成
		chromedp.Sleep(2*time.Second),
		chromedp.FullScreenshot(&buf, 100),
	); err != nil {
		log.Fatal(err)
	}
	// 将截图数据解码为图像
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		log.Fatal(err)
	}
	// 创建一个图像文件
	file, err := os.Create(fileName)
	log.Println("截图成功 创建图片", fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// 将图像保存为PNG文件
	err = png.Encode(file, img)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Image saved to output.png")
}

func HandleRiBao() string {
	// 获取今天的日期
	today := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("./screenLock-%s.png", today)
	println("获取今日日报图片" + fileName)
	// 判断是否有文件，如果有就 直接返回数据
	if _, err := os.Stat(fileName); err == nil {
		fmt.Println("fileName has exist , return it")
		return fileName
	}
	// 如果没有就开始启动截图
	// 启动一个 http 服务 用来提供 html 渲染
	go handleWebServe(false)
	// 等待服务启动
	screenLock(fileName)
	return fileName
}

func corn() {
	log.Printf("开启自动生成可莉日报任务")
	s := gocron.NewScheduler()
	err := s.Every(1).Days().At("00:01").Do(func() {
		log.Printf("执行自动生成可莉日报任务")
		// 清除本地所有的 生成文件
		filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
			if strings.Contains(path, "screenLock-") {
				os.Remove(path)
			}
			return nil
		})
		// 获取今天的日期
		today := time.Now().Format("2006-01-02")
		fileName := fmt.Sprintf("screenLock-%s.png", today)
		if _, err := os.Stat(fileName); err == nil {
			os.Remove(fileName)
			fmt.Println("fileName has exist , return it")
		}
		screenLock(fileName)
	})
	if err != nil {
		return
	}
	<-s.Start()
}

func main1() {
	go corn()
	handleWebServe(true)
}
