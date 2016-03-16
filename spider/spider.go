package spider

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

import (
	"code.google.com/p/mahonia"
	gxhtml "golang.org/x/net/html"
	bdurl "www.baidu.com/golang-lib/net/url"
)

import (
	"miniSpider/config"
	"miniSpider/log"
	"miniSpider/queue"
)

const (
	MAX_FILE_NAME_LENGTH int    = 255
	MAP_MD5_FILE_NAME    string = "url_md5_map"
)

type Spider struct {
	Config        config.ConfigStruct
	Queue         *queue.Queue
	ServerMap     map[string]*queue.ServerMap
	CrawlTimeout  time.Duration
	CrawlInterval time.Duration
}

func NewSpider(conf config.ConfigStruct) *Spider {
	interval := time.Duration(conf.Spider.CrawlInterval) * time.Second
	timeout := time.Duration(conf.Spider.CrawlTimeout) * time.Second

	return &Spider{
		Config:        conf,
		Queue:         queue.NewQueue(),
		ServerMap:     make(map[string]*queue.ServerMap),
		CrawlTimeout:  interval,
		CrawlInterval: timeout,
	}
}

//将种子文件读取到队列调度器中
func (s *Spider) Init() error {
	//读取种子文件
	seedJson, err := config.ReadJsonFile(s.Config.Spider.UrlListFile)
	if err != nil {
		return err
	}

	//插入种子目录地址
	for _, v := range seedJson {
		s.Queue.Push(config.CrawlInfo{Url: v, Depth: 0})
	}

	return nil
}

//抓取url网页内容
func (s *Spider) fetch(url string) string {
	//监听站点
	s.hostMonitor(url)

	var timeout = time.Duration(s.CrawlTimeout)

	tr := &http.Transport{
		//使用带超时的连接函数
		Dial: func(network, addr string) (net.Conn, error) {
			deadline := time.Now().Add(2 * timeout * time.Second)
			c, err := net.DialTimeout(network, addr, timeout)
			if err != nil {
				log.Warn(fmt.Sprintf("连接超时 url [%s]", url))
				return nil, err
			}
			c.SetDeadline(deadline)
			return c, nil
		},
		//建立连接后读超时
		ResponseHeaderTimeout: 2 * timeout * time.Second,
	}
	client := &http.Client{
		Transport: tr,
		//总超时，包含连接读写
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Warn(fmt.Sprintf("fetch url [%s] failed : %s", url, err))
		return ""
	}
	req.Header.Set("Connection", "keep-alive")
	//	req.Header.Set("Connection", "close")
	resp, err := client.Do(req)
	if err != nil {
		log.Warn(fmt.Sprintf("fetch url [%s] failed : %s", url, err))
		return ""
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn(fmt.Sprintf("fetch url [%s] content empty : %s", url, err))
		return ""
	}

	//gbk转utf8
	bodyString := string(body)
	enc := mahonia.NewEncoder("UTF-8")
	bodyString = enc.ConvertString(bodyString)

	return bodyString
}

//匹配正则，查找中符合条件url
func (s *Spider) regexpUrls(url string, ctx string, depth int) ([]string, error) {
	urlre, err := regexp.Compile(s.Config.Spider.TargetUrl)
	if err != nil {
		log.Warn(fmt.Sprintf("failed to compile regexp: %s", err))
		return nil, err
	}

	var urls []string
	var vurl string

	//处理页面页面链接
	doc, err := gxhtml.Parse(strings.NewReader(ctx))
	if err != nil {
		log.Warn(fmt.Sprintf("failed to parse content: %s", err))
		return nil, err
	}
	var f func(*gxhtml.Node)
	f = func(n *gxhtml.Node) {
		if n.Type == gxhtml.ElementNode {
			for _, a := range n.Attr {
				if a.Key == "src" || a.Key == "href" {
					matched, err := regexp.MatchString("javascript:location", a.Val)
					if err == nil && matched {
						continue
					}

					//符合url正则
					err2 := urlre.MatchString(a.Val)
					if !err2 {
						continue
					}

					//处理相对/绝对路径
					vurl = s.urlResolveReference(a.Val, url)
					if vurl != "" {
						urls = append(urls, vurl)
						log.Info(fmt.Sprintf("the sub url is: %s", vurl))
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	//放入队列调度器中
	for _, vurl := range urls {
		s.Queue.Push(config.CrawlInfo{Url: vurl, Depth: depth + 1})
	}

	return urls, nil
}

//整合抓取
func (s *Spider) Crawl(url string, depth int, done chan<- int) {
	log.Info(fmt.Sprintf("抓取url[%s]-depth[%d] 开始", url, depth))

	ctx := s.fetch(url)
	if ctx != "" {
		//根据内容生成文件
		go s.createHtml(url, ctx, depth)

		//在最大抓取深度之下，将页面符合要求记录添加到队列中
		if s.Config.Spider.MaxDepth > depth {
			//处理抓取页面包含url
			go s.regexpUrls(url, ctx, depth)
		}
	} else {
		log.Info(fmt.Sprintf("抓取url[%s]-depth[%d]失败", url, depth))
	}

	done <- 1
}

//转义url，使其可当文件名
func (s *Spider) urlEscape(url string) string {
	filename := bdurl.QueryEscape(url)

	encode := false

	if len(filename) > MAX_FILE_NAME_LENGTH {
		bits := md5.Sum([]byte(filename))
		filename = fmt.Sprintf("%x", string(bits[:]))
		encode = true
	}

	//需要写入映射文件中
	if encode {
		go s.appendMd5Map(url, filename)
	}

	return filename
}

//生成文件
func (s *Spider) createHtml(url string, ctx string, depth int) bool {
	//转义url
	filename := s.urlEscape(url)

	filename = s.Config.Spider.OutputDirectory + "/" + filename
	op := false

	if s.checkFileIsExist(filename) {
		err := os.Remove(filename)
		if err != nil {
			log.Warn(fmt.Sprintf("failed to remove file[%s]: %s", filename, err))
			return false
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Warn(fmt.Sprintf("failed to open file[%s]: %s", filename, err))
		return false
	}
	defer file.Close()

	_, err2 := file.WriteString(ctx + "\r\n")
	if err2 != nil {
		log.Warn(fmt.Sprintf("failed to writeString file[%s]: %s", filename, err))
		return false
	}

	op = true

	if op {
		log.Info(fmt.Sprintf("抓取url[%s]-depth[%d]成功", url, depth))
	} else {
		log.Info(fmt.Sprintf("抓取url[%s]-depth[%d]失败", url, depth))
	}

	return op
}

//映射文件名
func (s *Spider) appendMd5Map(url string, filename string) error {
	mapfile := s.Config.Spider.OutputDirectory + "/" + MAP_MD5_FILE_NAME
	f, err := os.OpenFile(mapfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Warn(fmt.Sprintf("write md5map failed url[%s] - filename[%s]: %s", url, filename, err))
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s %s\n", filename, url)

	log.Info(fmt.Sprintf("write md5map result url[%s] - filename[%s]: %s", url, filename, err))

	return err
}

//获取url绝对地址
func (s *Spider) urlResolveReference(relUrl string, baseUrl string) string {
	rel, err := url.Parse(relUrl)
	if err != nil {
		log.Warn(fmt.Sprintf("resolveReference url[%s]: %s", relUrl, err))
		return ""
	}

	base, err := url.Parse(baseUrl)
	if err != nil {
		log.Warn(fmt.Sprintf("resolveReference url[%s]: %s", baseUrl, err))
		return ""
	}

	return base.ResolveReference(rel).String()
}

//判断文件是否存在
func (s *Spider) checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

//站点监听器
func (s *Spider) hostMonitor(uurl string) {
	v, err := url.Parse(uurl)
	if err != nil {
		log.Warn(fmt.Sprintf("hostMonitor url[%s]: %s", uurl, err))
	} else {
		host := v.Host
		t, exists := s.ServerMap[host]
		if !exists {
			t = queue.NewServerMap(host, s.CrawlInterval, s.CrawlTimeout)
			t.SetLastOpTime(time.Now())
			t.AddCount()
			s.ServerMap[host] = t
		} else {
			t.AddCount()
			t.Sleep()

			if time.Since(t.LastOpTime()) > t.Timeout() {
				delete(s.ServerMap, host)
			}
		}
	}
}
