package btsite

import (
	"fmt"
	"github.com/heibizi/go-siteadapt"
	"net/url"
	"strings"
)

type (
	// Client 站点客户端
	Client interface {
		// Favicon 获取站点 favicon 文件
		Favicon() ([]byte, error)
		// UserBasicInfo 获取用户基础信息，通常是首页能拿到的
		UserBasicInfo() (UserBasicInfo, error)
		// UserDetails 获取用户详情，通常是详情页能拿到的
		UserDetails() (UserDetails, error)
		// Search 搜索种子列表
		Search(searchParams SearchParams) ([]SearchTorrent, error)
		// SeedingStatistics 获取做种统计信息
		SeedingStatistics() (SeedingStatistics, error)
		// MyHr HR 考核中列表
		MyHr() ([]HrTorrent, error)
		// UnreadMessages 未读消息列表，只取第一页，bool：是否跳转到详情
		UnreadMessages(detail bool) ([]Message, error)
		// LatestNotice 最新公告
		LatestNotice() (*Notice, error)
		// Rss RSS 拉取
		Rss() ([]RssTorrent, error)
		// SignIn 签到
		SignIn() (SignInResult, error)
		// GetDownloadUrl 获取种子下载地址
		GetDownloadUrl(torrent SearchTorrent) (string, error)
		// Details 获取种子详情
		Details(id string) (TorrentDetail, error)
	}
	// requestSiteParams 站点请求参数
	// 自定义请求的优先级：reqId > rd > schema
	requestSiteParams struct {
		site     *Site                        // 站点
		reqId    requestId                    // 请求 id
		rd       *siteadapt.RequestDefinition // 自定义请求
		path     string                       // 替换 rd 的 path
		params   url.Values                   // url 请求参数
		formData url.Values                   // form-data 请求参数
		env      map[string]string            // 环境变量
		body     map[string]any               // 请求体
	}
)

// 注册 Client 实现类的构造函数映射
var clientFactoryRegistry = map[string]func(*Site) Client{
	string(siteSchemaNexusPHP): func(site *Site) Client { return &npClient{site} },
	string(siteSchemaMTorrent): func(site *Site) Client {
		return &mtClient{site, &npClient{site}}
	},
}

// NewClient 根据站点的唯一标识获取站点配置，根据站点配置的系统架构类型创建客户端
func NewClient(site *Site) (Client, error) {
	sc, err := SiteHelper.GetConfigByCode(site.Code)
	if err != nil {
		return nil, err
	}
	if constructor, ok := clientFactoryRegistry[sc.Schema]; ok {
		return constructor(site), nil
	}
	return nil, fmt.Errorf("无效架构: %s", sc.Schema)
}

func newSiteAdapt(params requestSiteParams) (*siteadapt.SiteAdaptor, *siteadapt.RequestSiteParams, error) {
	sh := SiteHelper
	site := params.site
	sc, err := sh.GetConfigByCode(site.Code)
	if err != nil {
		return nil, nil, err
	}
	env := params.env
	if env == nil {
		env = make(map[string]string)
	}
	// 常用变量
	env["userId"] = site.UserId
	env["api"] = site.Api
	if site.Domain != "" {
		env["domain"] = site.Domain
	}
	rsp := siteadapt.RequestSiteParams{
		ReqId:    string(params.reqId),
		Rd:       params.rd,
		Domain:   site.Domain,
		Api:      site.Api,
		Path:     params.path,
		Params:   params.params,
		FormData: params.formData,
		Body:     params.body,
		Env:      env,
		UA:       site.UserAgent,
		Cookie:   site.Cookie,
	}
	// 自定义请求头
	if len(site.Headers) > 0 {
		headers := make(map[string]string)
		for _, header := range strings.Split(site.Headers, "\n") {
			kvs := strings.Split(strings.TrimSpace(header), ":")
			if len(kvs) == 2 {
				headers[kvs[0]] = kvs[1]
			}
		}
		rsp.Headers = headers
	}
	return siteadapt.NewSiteAdaptor(sc.Config), &rsp, nil
}

// data 获取对象数据
func data(params requestSiteParams, output any, fn siteadapt.DataFunc) error {
	sa, rsp, err := newSiteAdapt(params)
	if err != nil {
		return err
	}
	err = sa.Data(*rsp, output, fn)
	if err != nil {
		return err
	}
	return nil
}

// list 获取列表数据
func list(params requestSiteParams, output any, fn siteadapt.ListFunc) error {
	sa, rsp, err := newSiteAdapt(params)
	if err != nil {
		return err
	}
	err = sa.List(*rsp, output, fn)
	if err != nil {
		return err
	}
	return nil
}

// raw 获取原始数据
func raw(params requestSiteParams, fn siteadapt.RawFunc) error {
	sa, rsp, err := newSiteAdapt(params)
	if err != nil {
		return err
	}
	err = sa.Raw(*rsp, fn)
	if err != nil {
		return err
	}
	return nil
}

func newError(site *Site, err error, format string, v ...any) error {
	return fmt.Errorf("站点(%s)%s, 异常: %v", site.Name, fmt.Sprintf(format, v...), err)
}
