package main

import (
	"bufio"
	"fmt"
	"github.com/gocolly/colly"
	"io"
	"net/http"
	"os"
)

var fileNum int
func main(){
	findImg("https://www.zcool.com.cn/work/ZMzU5MzQxODg=.html")
}

func findImg(urlstr string){
	// 创建采集器
	c1 := colly.NewCollector(colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.190 Safari/537.36"), colly.MaxDepth(1))
	// 处理返回的html
	c1.OnHTML("#lightbox-img-wrap > ul", func(e *colly.HTMLElement) {
		e.ForEach("li", func(i int, item *colly.HTMLElement) {
			href := item.ChildAttr("div[class='light-slide-content'] > img", "data-src")
			downImg(href)
		})
	})
	// 在向服务器发送请求前先打印一下（发送请求前的回调）
	c1.OnRequest(func(r *colly.Request) {
		fmt.Println("爬取页面：", r.URL)
	})
	// 出现错误时的日志打印
	c1.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})
	//向url站点发送请求
	err := c1.Visit(urlstr)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func downImg(imgUrl string){
	dir := "./picTest"
	makeDir(dir)
	fileNum++
	fileName := fmt.Sprintf("%s/%d.png",dir,fileNum)

	res, err := http.Get(imgUrl)
	if err != nil {
		fmt.Println("A error occurred!")
		return
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

// 检查文件存储路径如果文件夹不存在则创建
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

