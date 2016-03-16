package miniSpider

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/pprof"
)

import (
	"miniSpider/config"
	"miniSpider/log"
	"miniSpider/spider"
)

func init() {
	//使用多核,1.5以上版本默认开启,忽略
	//	fmt.Printf("%s", runtime.Version())
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	p := pprof.Lookup("goroutine")
	p.WriteTo(w, 1)
}

func NewSpider() {
	//初始化
	config.NewInit()

	defer func() {
		if err := recover(); err != nil {
			log.Warn(fmt.Sprintf("panic error: [%s]", err))
			config.Exit()
		}
	}()

	go func() {
		http.HandleFunc("/", handler)
		http.ListenAndServe(":11181", nil)
	}()

	//初始化配置信息
	cfg, err := config.NewConf()
	if err != nil {
		log.Warn(fmt.Sprintf("配置信息错误, err:[%s]", err))
		config.Exit()
	}

	//初始化爬虫
	ns := spider.NewSpider(cfg)
	if err := ns.Init(); err != nil {
		log.Warn(fmt.Sprintf("spider init err:[%s]", err))
		config.Exit()
	}

	//使用routine并行数量
	ch := make(chan int, ns.Config.Spider.ThreadCount)
	var crawlInfo config.CrawlInfo

	for {
		num := ns.Queue.Len()
		if num > 0 {
			crawlInfo = ns.Queue.Pop().(config.CrawlInfo)
			if crawlInfo.Url != "" {
				go func(url string, depth int) {
					Task(url, depth, ch, ns)
				}(crawlInfo.Url, crawlInfo.Depth)

				<-ch
			} else {
				continue
			}
		}

		if ns.Queue.Empty() {
			goto Exit
		}
	}

Exit:
	config.Exit()
}

func Task(url string, depth int, ch chan int, ns *spider.Spider) {
	//使用routine并行数量
	nch := make(chan<- int, ns.Config.Spider.ThreadCount)
	ns.Crawl(url, depth, nch)

	ch <- 1
}
