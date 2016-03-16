package spider

import (
	"os"
	"testing"
)

import (
	"miniSpider/config"
)

var cfg config.ConfigStruct
var ns = NewSpider(cfg)

func Test_createHtml(t *testing.T) {
	os.MkdirAll("D:/SharedFolder/go/pro/temp/html/", 0777)

	var tests = []struct {
		filename string
		want     bool
	}{
		{"D:/SharedFolder/go/pro/temp/html/1.html", true},
		{"D:/SharedFolder/go/pro/temp/html/2.html", true},
		{"D:/SharedFolder/go/pro/temp/html/3.html", true},
	}
	for _, test := range tests {
		if got := ns.createHtml(test.filename, "内容", 1); got != test.want {
			t.Errorf("createHtml(%q) = %v", test.filename, got)
		}
	}
}

func Test_urlResolveReference(t *testing.T) {
	var tests = []struct {
		base, rel, want string
	}{
		{"http://foo.com/bar", "/baz", "http://foo.com/baz"},
		{"http://foo.com/bar?a=b#f", "/baz", "http://foo.com/baz"},
		{"http://foo.com/bar?a=b", "/baz?c=d", "http://foo.com/baz?c=d"},
		{"http://foo.com/bar/baz", "../quux", "http://foo.com/quux"},
		{"http://foo.com/bar/baz", "../../../../../quux", "http://foo.com/quux"},
		{"http://foo.com/bar", "..", "http://foo.com/"},
		{"http://foo.com/bar/baz", "./..", "http://foo.com/"},
	}

	for _, test := range tests {
		if got := ns.urlResolveReference(test.rel, test.base); got != test.want {
			t.Errorf("urlResolveReference(%s, %s) = %s", test.rel, test.base, got)
		}
	}
}

func Test_checkFileIsExist(t *testing.T) {
	var tests = []struct {
		filename string
		want     bool
	}{
		{"D:/SharedFolder/go/pro/temp/html/1.html", true},
		{"D:/SharedFolder/go/pro/temp/html/2.html", true},
		{"D:/SharedFolder/go/pro/temp/html/3.html", true},
		{"D:/SharedFolder/go/pro/temp/html/4.html", false},
	}
	for _, test := range tests {
		if got := ns.checkFileIsExist(test.filename); got != test.want {
			t.Errorf("checkFileIsExist(%q) = %v", test.filename, got)
		}
	}
}
