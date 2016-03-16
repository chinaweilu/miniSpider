package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

import (
	//	"gopkg.in/gcfg.v1"
	"code.google.com/p/gcfg"
	bdlog "www.baidu.com/golang-lib/log"
)

var (
	Version    string = "1.0"
	Author     string = "baidu"
	Configfile string
	Logfile    string
)

type ConfigStruct struct {
	Spider struct {
		UrlListFile     string
		OutputDirectory string
		MaxDepth        int
		CrawlInterval   int
		CrawlTimeout    int
		TargetUrl       string
		ThreadCount     int
	}
}

type CrawlInfo struct {
	Url   string
	Depth int
	//Done  chan<- bool
}

func NewInit() {
	//D:/SharedFolder/go/pro/src/miniSpider/conf/spider.conf
	//D:/SharedFolder/go/pro/src/miniSpider/data/logs
	flag.StringVar(&Configfile, "c", "", "配置文件路径")
	flag.StringVar(&Logfile, "l", "", "日志文件输出目录")

	if len(os.Args) > 1 {
		args := os.Args[1]
		switch string(args) {
		case "-v", "-version":
			fmt.Println("miniSpider Version", Version)
			os.Exit(0)
		case "-h", "-help":
			flag.Usage = Usage
			flag.Parse()
			os.Exit(0)
		}
	}

	flag.Usage = Usage
	flag.Parse()

	if Configfile == "" {
		fmt.Println("配置文件路径不能为空")
		os.Exit(0)
	} else if !checkFileIsExist(Configfile) {
		fmt.Println("配置文件路径错误")
		os.Exit(0)
	}

	if Logfile == "" {
		fmt.Println("日志文件输出目录不能为空")
		os.Exit(0)
	} else {
		//		if !IsDir(Logfile) {
		//			fmt.Println("日志文件输出目录出错")
		//			os.Exit(0)
		//		}
		NewLog()
	}
}

func NewConf() (ConfigStruct, error) {

	if Configfile == "" {
		return ConfigStruct{}, nil
	}
	if !checkFileIsExist(Configfile) {
		return ConfigStruct{}, nil
	}

	var cfg ConfigStruct
	err := gcfg.ReadFileInto(&cfg, Configfile)
	if err != nil {
		NewLog()
		bdlog.Logger.Error(fmt.Sprintf("配置文件错误 %s", err))

		return ConfigStruct{}, nil
	}

	//创建输出目录
	os.MkdirAll(cfg.Spider.OutputDirectory, 0777)

	return cfg, nil
}

func NewLog() bool {
	if Logfile == "" {
		return false
	}
	//	if !IsDir(Logfile) {
	//		return false
	//	}

	//创建输出目录
	os.MkdirAll(Logfile, 0777)

	bdlog.Init("mini_spider", "INFO", Logfile, true, "midnight", 5)

	return true
}

func ReadJsonFile(filename string) ([]string, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		bdlog.Logger.Warn(fmt.Sprintf("readJsonFile err: %s ", err))
		return nil, err
	}

	var seedJson []string
	if err := json.Unmarshal(bytes, &seedJson); err != nil {
		bdlog.Logger.Warn(fmt.Sprintf("unmarshal err: %s", err))
		return nil, err
	}

	return seedJson, nil
}

// 帮助信息
func Usage() {
	fmt.Fprintln(os.Stderr, `用法：
mini_spider [-c </spider.conf>] [-l </data/logs/>] [-v] [-h]
            `)
	fmt.Fprintln(os.Stderr, "\n选项说明")
	fmt.Fprintf(os.Stderr, "\n  %-5s %-5s %s\n", "选项", "默认值", "说明")
	flag.VisitAll(func(f *flag.Flag) {
		if f.Value.String() == "" {
			fmt.Fprintf(os.Stderr, "  -%-6s %-7s %s\n", f.Name, "", f.Usage)
		} else {
			fmt.Fprintf(os.Stderr, "  -%-6s %-8s %s\n", f.Name, f.DefValue, f.Usage)
		}
	})
	fmt.Fprintf(os.Stderr, "  -%-6s %-7s %s\n", "v", "--", "版本")
	fmt.Fprintf(os.Stderr, "  -%-6s %-7s %s\n", "h", "--", "帮助说明")
	fmt.Fprintf(os.Stderr, "\n作者：%s\n", Author)
}

func Exit() {
	fmt.Println("程序结束,谢谢再见")
	os.Exit(0)
}

//判断文件是否存在
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// 判断给定文件名是否是一个目录
// 如果文件名存在并且为目录则返回 true。如果 filename 是一个相对路径，则按照当前工作目录检查其相对路径。
func IsDir(filename string) bool {
	return isFileOrDir(filename, true)
}

// 判断给定文件名是否为一个正常的文件
// 如果文件存在且为正常的文件则返回 true
func IsFile(filename string) bool {
	return isFileOrDir(filename, false)
}

// 判断是文件还是目录，根据decideDir为true表示判断是否为目录；否则判断是否为文件
func isFileOrDir(filename string, decideDir bool) bool {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return false
	}
	isDir := fileInfo.IsDir()
	if decideDir {
		return isDir
	}
	return !isDir
}
