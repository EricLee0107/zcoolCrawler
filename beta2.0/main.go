// Command click2 is a chromedp example demonstrating how to use a selector to
// click on an element.
package main

import (
	"bufio"
	"fmt"
	"github.com/Unknwon/goconfig"
	"github.com/gocolly/colly"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)



type zcool struct{
	// 图片地址通道
	articleChannel chan string
	// 文章地址通道
	imgChannel chan string
	// 文章获取线程关闭通知通道
	goStopChannel chan struct{}
	stopFindGoNum int
	wait sync.WaitGroup
}

var CoZCool *zcool


var Cfg *goconfig.ConfigFile

// 存储路径
var downLoadDir string
// 图片存储文件夹名称
var dirName string
// 获取图片的goroutine数量
var findGoNum int
// 下载图片的goroutine数量
var downGoNum int





func init(){
	var err error
	Cfg,err = goconfig.LoadConfigFile("./beta2.0/config/conf.ini")
	if err != nil{
		log.Fatal(err)
	}
	downLoadDir,err = Cfg.GetValue(goconfig.DEFAULT_SECTION,"dir_path")
	if err != nil{
		log.Fatal(err)
	}
	if downLoadDir == ""{
		downLoadDir = "."
	}
	if downLoadDir[len(downLoadDir)-1] == '\\' || downLoadDir[len(downLoadDir)-1] == '/'{
		downLoadDir = downLoadDir[:len(downLoadDir)-1]
	}
	dirName,err = Cfg.GetValue(goconfig.DEFAULT_SECTION,"dir_name")
	if err != nil{
		log.Fatal(err)
	}
	findNumStr,err := Cfg.GetValue(goconfig.DEFAULT_SECTION,"find_go_num")
	if err != nil{
		log.Fatal(err)
	}

	findGoNum,err= strconv.Atoi(findNumStr)
	if err != nil{
		log.Fatal(err)
	}
	// 最小为1
	if findGoNum < 1{
		findGoNum = 1
	}
	downGoStr,err := Cfg.GetValue(goconfig.DEFAULT_SECTION,"down_go_num")
	if err != nil{
		log.Fatal(err)
	}

	downGoNum,err= strconv.Atoi(downGoStr)
	if err != nil{
		log.Fatal(err)
	}
	// 最小为1
	if downGoNum < 1{
		downGoNum = 1
	}
}


func main() {
	CoZCool = new(zcool)
	CoZCool.articleChannel = make(chan string,1000)
	CoZCool.goStopChannel = make(chan struct{})
	CoZCool.imgChannel = make(chan string, 1000)
	// 根据获取到的收藏夹连接遍历所有的收藏夹
	go func(){
		CoZCool.wait.Add(1)
		urls := readUrls()
		for _,url:=range urls{
			if url == ""{
				continue
			}
			CoZCool.articleChannel <- url
		}
		close(CoZCool.articleChannel)
		CoZCool.wait.Done()
	}()
	// 启动N个go去获取图片地址
	for i:=0;i<findGoNum;i++{
		go findImg()
	}
	// 创建文件夹
	dir := fmt.Sprintf("%s/%s",downLoadDir,dirName)
	makeDir(dir)
	// 启动M个go去下载图片
	for k:=0;k<downGoNum;k++{
		go downImg(dir,k)
	}
	go stopImgChan()
	CoZCool.wait.Wait()
	fmt.Println("获取完成！")
}

func readUrls () []string{
	fi, err := os.Open("./beta2.0/config/urls.conf")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return nil
	}
	defer fi.Close()

	urls := make([]string,0)
	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		urls = append(urls,string(a))
	}
	return urls
}


func downImg(dir string, startNum int){
	CoZCool.wait.Add(1)
	defer CoZCool.wait.Done()
	subDir := fmt.Sprintf("%s/%d",dir,startNum)
	fileNum := 0
	makeDir(subDir)
	for {
		select{
		case imgUrl,ok := <- CoZCool.imgChannel:
			if !ok{
				return
			}
			fileNum++
			fileName := fmt.Sprintf("%s/%s-%d.png",subDir,dirName,fileNum)

			res, err := http.Get(imgUrl)
			if err != nil {
				fmt.Println("A error occurred!")
				break
			}
			defer res.Body.Close()
			// 获得get请求响应的reader对象
			reader := bufio.NewReaderSize(res.Body, 64 * 1024)


			file, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			// 获得文件的writer对象
			writer := bufio.NewWriter(file)

			io.Copy(writer, reader)
			fmt.Printf("write:%s\n", fileName)
		}
	}
}

func stopImgChan(){
	CoZCool.wait.Add(1)
	defer CoZCool.wait.Done()
	for {
		select{
		case <-CoZCool.goStopChannel:
			CoZCool.stopFindGoNum ++
			if CoZCool.stopFindGoNum == findGoNum{
				close(CoZCool.imgChannel)
			}
		}
	}

}

func findImg(){
	CoZCool.wait.Add(1)
	defer CoZCool.wait.Done()
	for {
		select {
		case urlstr,ok :=<- CoZCool.articleChannel:
			// 关闭go
			if !ok{
				CoZCool.goStopChannel <- struct{}{}
				return
			}
			c1 := colly.NewCollector(colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.190 Safari/537.36"), colly.MaxDepth(1))
			//采集器1，获取文章列表
			c1.OnHTML("#lightbox-img-wrap > ul", func(e *colly.HTMLElement) {
				e.ForEach("li", func(i int, item *colly.HTMLElement) {
					href := item.ChildAttr("div[class='light-slide-content'] > img", "data-src")
					CoZCool.imgChannel <- href
				})
			})

			c1.OnRequest(func(r *colly.Request) {
				fmt.Println("c1爬取页面：", r.URL)
			})

			c1.OnError(func(r *colly.Response, err error) {
				fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
			})

			err := c1.Visit(urlstr)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}

}



// 判断文件夹是否存在
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func makeDir(path string)(bool,error){
	exist, err := pathExists(path)
	if err != nil {
		return false,err
	}
	if !exist{
		// 创建文件夹
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return false,err
		} else {
			return true,nil
		}
	}
	return true,nil
}


