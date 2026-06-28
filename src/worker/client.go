package worker

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	userAgent  string
}

func (c *Client) get(url string) (*http.Response, error) {
	return c.getConditional(url, "", "")
}

func (c *Client) getConditional(url, lastModified, etag string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	return c.httpClient.Do(req)
}

var client *Client

// 固化代理配置 - 请根据需要修改这些值
const proxyURL = "http://127.0.0.1:17890"                        // 替换为你的代理地址
var domainList = []string{"rsshub.rssforever.com", "github.com"} // 替换为需要使用代理的域名列表

// 代理函数：匹配域名列表时使用指定代理，否则使用环境变量代理
func getProxyFunc(proxyURL string, domains []string) func(*http.Request) (*url.URL, error) {
	return func(req *http.Request) (*url.URL, error) {
		for _, domain := range domains {
			if strings.Contains(req.URL.Host, domain) {
				return url.Parse(proxyURL)
			}
		}
		// 不匹配时使用环境变量代理
		return http.ProxyFromEnvironment(req)
	}
}

func SetVersion(num string) {
	client.userAgent = "Yarr/" + num
}

func init() {
	transport := &http.Transport{
		Proxy: getProxyFunc(proxyURL, domainList),
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).DialContext,
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: time.Second * 10,
	}
	httpClient := &http.Client{
		Timeout:   time.Second * 30,
		Transport: transport,
	}
	client = &Client{
		httpClient: httpClient,
		userAgent:  "Yarr/1.0",
	}
}
