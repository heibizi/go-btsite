package btsite

import (
	"encoding/xml"
	"fmt"
	"github.com/heibizi/go-siteadapt"
	"net/url"
	"strings"
	"time"
)

type (
	// npClient NexusPHP 客户端
	npClient struct {
		site *Site
	}
)

func (c *npClient) Favicon() ([]byte, error) {
	var data []byte
	err := raw(requestSiteParams{
		site:  c.site,
		reqId: requestIdFavicon,
	}, func(result siteadapt.RawResult) {
		data = result.Data
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *npClient) UserBasicInfo() (UserBasicInfo, error) {
	var ud UserBasicInfo
	err := data(requestSiteParams{
		site:  c.site,
		reqId: requestIdUserBasicInfo,
	}, &ud, nil)
	if err != nil {
		return ud, newError(c.site, err, "解析基础信息失败")
	}
	if ud.Ratio == 0 && ud.Downloaded > 0 {
		ratio := float64(10 ^ 3)
		ud.Ratio = float64(int(float64(ud.Uploaded)/float64(ud.Downloaded)*ratio)) / ratio
	}
	if ud.Bonus == 0 && (ud.Gold > 0 || ud.Silver > 0 || ud.Copper > 0) {
		totalCopper := ud.Gold*100*100 + ud.Silver*100 + ud.Copper
		ud.Bonus = totalCopper
	}
	sc, err := SiteHelper.GetConfigByCode(c.site.Code)
	if err != nil {
		return ud, err
	}
	if sc.CountMessage {
		messages, err := c.UnreadMessages(false)
		if err != nil {
			return ud, err
		}
		ud.UnreadMessageCount = len(messages)
	}
	return ud, nil
}

func (c *npClient) UserDetails() (UserDetails, error) {
	var ud UserDetails
	err := data(requestSiteParams{
		site:  c.site,
		reqId: requestIdUserDetails,
	}, &ud, nil)
	if err != nil {
		return ud, newError(c.site, err, "解析用户详情信息异常")
	}
	return ud, nil
}

func (c *npClient) Search(searchParams SearchParams) ([]SearchTorrent, error) {
	sh := SiteHelper
	site := c.site
	sc, err := sh.GetConfigByCode(site.Code)
	if err != nil {
		return nil, newError(site, err, "未获取到站点配置")
	}
	var params url.Values
	var paramsEnv = map[string]string{
		"keyword": searchParams.Keyword,
	}
	if len(searchParams.Keyword) > 0 {
		params = url.Values{
			"search_mode": {"0"},
			"page":        {fmt.Sprintf("%d", searchParams.Page)},
			"notnewword":  {"1"},
		}
		category := &sc.Categories
		if category != nil {
			var cats []AdaptMediaCat
			switch searchParams.MediaType {
			case Movie:
				cats = category.Movie
			case Tv:
				cats = category.TV
			default:
				cats = append(category.Movie, category.TV...)
			}
			for _, cat := range cats {
				field := category.Field
				if len(field) > 0 {
					value := params.Get(field)
					params.Add(field, value+category.Delimiter+cat.ID)
				} else {
					params.Add(cat.ID, "1")
				}
			}
		}
	} else {
		params = url.Values{
			"page": {fmt.Sprintf("%d", searchParams.Page)},
		}
	}
	var torrents []torrent
	requestUrl := ""
	err = data(requestSiteParams{
		site:   site,
		reqId:  requestIdSearch,
		params: params,
		env:    paramsEnv,
	}, &torrents, func(result siteadapt.DataResult) {
		requestUrl = result.RequestUrl
	})
	if err != nil {
		return nil, newError(site, err, "解析种子列表失败")
	}
	var searchTorrents []SearchTorrent
	for _, torrent := range torrents {
		var pageUrl string
		if strings.HasPrefix(torrent.Details, "http") {
			pageUrl = torrent.Details
		} else {
			detailLink, err := JoinURL(requestUrl, torrent.Details)
			if err != nil {
				return nil, newError(site, err, "搜索种子拼接 details 错误")
			}
			pageUrl = detailLink
		}
		var enclosure string
		if strings.HasPrefix(torrent.Download, "http") || strings.HasPrefix(torrent.Download, "magnet") {
			enclosure = torrent.Download
		} else {
			enclosureLink, err := JoinURL(requestUrl, torrent.Download)
			if err != nil {
				return nil, newError(site, err, "搜索种子解析 download 错误")
			}
			enclosure = enclosureLink
		}
		search := SearchTorrent{
			ID:                   torrent.ID,
			Category:             torrent.Category,
			Title:                torrent.Title,
			Description:          torrent.Description,
			PageURL:              pageUrl,
			Enclosure:            enclosure,
			Grabs:                torrent.Grabs,
			Leechers:             torrent.Leechers,
			Seeders:              torrent.Seeders,
			Size:                 torrent.Size,
			DownloadVolumeFactor: torrent.DownloadVolumeFactor,
			UploadVolumeFactor:   torrent.UploadVolumeFactor,
			PubDate:              torrent.DateAdded,
			DateElapsed:          torrent.DateElapsed,
			HrDays:               torrent.HrDays,
			HitAndRun:            torrent.HrDays > 0,
			Labels:               torrent.Labels,
		}
		searchTorrents = append(searchTorrents, search)
	}
	return searchTorrents, nil
}

func (c *npClient) SeedingStatistics() (SeedingStatistics, error) {
	sc, err := SiteHelper.GetConfigByCode(c.site.Code)
	if err != nil {
		return SeedingStatistics{}, err
	}
	rd, exists := sc.RequestDefinitions[string(requestIdSeedingStatistics)]
	// 如果定义了请求，且不走分页
	if exists && rd.List == nil {
		var ss SeedingStatistics
		err := data(requestSiteParams{
			site:  c.site,
			reqId: requestIdSeedingStatistics,
		}, &ss, nil)
		if err != nil {
			return ss, newError(c.site, err, "做种信息失败")
		}
		return ss, nil
	} else {
		// 分页统计做种信息
		var seedingList []seeding
		currentPageSeedingList, nextPage, err := c.CurrentPageSeeding("")
		if err != nil {
			return SeedingStatistics{}, err
		}
		seedingList = append(seedingList, currentPageSeedingList...)
		for len(nextPage) > 0 {
			currentPageSeedingList, nextPageTmp, err := c.CurrentPageSeeding(nextPage)
			if err != nil {
				return SeedingStatistics{}, err
			}
			seedingList = append(seedingList, currentPageSeedingList...)
			nextPage = nextPageTmp
			time.Sleep(500 * time.Millisecond)
		}
		var size int64 = 0
		for _, seeding := range seedingList {
			size = size + seeding.Size
		}
		return SeedingStatistics{
			Count: len(seedingList),
			Size:  size,
		}, nil
	}
}

// CurrentPageSeeding 当前页做种信息以及下一页链接地址
func (c *npClient) CurrentPageSeeding(url string) ([]seeding, string, error) {
	var seedingList []seeding
	nextPage := ""
	err := list(requestSiteParams{
		site:  c.site,
		reqId: requestIdSeedingStatistics,
		path:  url,
	}, &seedingList, func(result siteadapt.ListResult) {
		nextPage = result.NextPage
	})
	if err != nil {
		return nil, "", newError(c.site, err, "解析做种信息列表失败")
	}
	return seedingList, nextPage, nil
}

func (c *npClient) MyHr() ([]HrTorrent, error) {
	sc, err := SiteHelper.GetConfigByCode(c.site.Code)
	if err != nil {
		return nil, err
	}
	if !sc.Price.HasHR {
		return nil, nil
	}
	var hrList []HrTorrent
	err = list(requestSiteParams{
		site:  c.site,
		reqId: requestIdMyHr,
	}, &hrList, nil)
	if err != nil {
		return nil, newError(c.site, err, "HR列表失败")
	}
	return hrList, nil
}

func (c *npClient) UnreadMessages(detail bool) ([]Message, error) {
	var o []Message
	requestUrl := ""
	err := list(requestSiteParams{
		site:  c.site,
		reqId: requestIdUnreadMessages,
	}, &o, func(result siteadapt.ListResult) {
		requestUrl = result.NextPage
	})
	if err != nil {
		return nil, newError(c.site, err, "未读消息列表异常")
	}
	if detail {
		for i := range o {
			message := &o[i]
			detailUrl, err := JoinURL(requestUrl, message.Link)
			if err != nil {
				return nil, err
			}
			detail, err := c.unreadMessageDetail(detailUrl)
			if err != nil {
				return nil, err
			}
			message.Content = detail.Content
		}
	}
	return o, nil
}

func (c *npClient) LatestNotice() (*Notice, error) {
	var notice Notice
	err := data(requestSiteParams{
		site:  c.site,
		reqId: requestIdLatestNotice,
	}, &notice, nil)
	if err != nil {
		return nil, newError(c.site, err, "解析最近公告失败")
	}
	// todo 去除标题为空的公告
	if notice.Title == "" {
		return nil, nil
	}
	return &notice, nil
}

func (c *npClient) Rss() ([]RssTorrent, error) {
	rd := siteadapt.RequestDefinition{
		Parser: "None",
		Method: "GET",
		Path:   c.site.RssUrl,
	}
	var data []byte
	err := raw(requestSiteParams{
		site: c.site,
		rd:   &rd,
	}, func(result siteadapt.RawResult) {
		data = result.Data
	})
	if err != nil {
		return nil, newError(c.site, err, "获取 RSS 数据异常")
	}
	var rss rssResult
	err = xml.Unmarshal(data, &rss)
	if err != nil {
		return nil, newError(c.site, err, "解析 RSS xml 数据异常")
	}
	var torrents []RssTorrent
	for _, item := range rss.Items {
		if len(item.Title) == 0 {
			continue
		}
		// todo 月月标题特殊处理
		//if siteDomain != "" {
		// Placeholder for special title processing
		//}
		link := item.Link
		enclosure := item.Enclosure.URL
		if len(enclosure) == 0 && len(link) == 0 {
			continue
		}
		if len(enclosure) == 0 && len(link) > 0 {
			enclosure = link
			link = ""
		}
		torrents = append(torrents, RssTorrent{
			ID:          item.Guid,
			Title:       item.Title,
			Enclosure:   enclosure,
			Size:        siteadapt.ParseInt64(item.Enclosure.Length),
			Description: item.Description,
			Link:        item.Link,
			PubDate:     siteadapt.GetTimeStamp(item.PubDate),
		})
	}
	return torrents, nil
}

// unreadMessageDetail 未读消息详情
func (c *npClient) unreadMessageDetail(url string) (Message, error) {
	var message Message
	err := data(requestSiteParams{
		site:  c.site,
		reqId: requestIdUnreadMessageDetail,
		path:  url,
	}, &message, nil)
	if err != nil {
		return message, newError(c.site, err, "用户未读消息详情失败")
	}
	return message, nil
}

func (c *npClient) SignIn() (SignInResult, error) {
	// 尝试获取用户基础信息，既可以判断是否需要已登录也可以用于模拟登录
	ubi, err := c.UserBasicInfo()
	if err != nil {
		return SignInResult{}, err
	}
	if !ubi.IsLogin {
		return SignInResult{
			Code:    SignInCodeNeedLogin,
			Message: "未登录",
		}, nil
	}
	if ubi.SignedIn {
		return SignInResult{
			Code:    SignInCodeSigned,
			Message: "今日已签到",
		}, nil
	}
	sc, err := SiteHelper.GetConfigByCode(c.site.Code)
	if err != nil {
		return SignInResult{}, err
	}
	// 无需签到
	if !sc.Required.SignIn {
		return SignInResult{
			Code:    SignInCodeSuccess,
			Message: "模拟登录成功",
		}, nil
	}
	// 签到
	r := signInResult{}
	statusCode := 0
	err = data(requestSiteParams{
		site:  c.site,
		reqId: requestIdSignIn,
	}, &r, func(result siteadapt.DataResult) {
		statusCode = result.StatusCode
	})
	if err != nil {
		return SignInResult{}, newError(c.site, err, "签到异常")
	}
	// 签到成功
	if r.SignedIn {
		return SignInResult{
			Code:    SignInCodeSuccess,
			Message: "签到成功",
		}, nil
	}
	// 签到失败
	if statusCode == 200 {
		return SignInResult{
			Code:    SignInCodeFailure,
			Message: fmt.Sprintf("签到失败，请检查该站点是否已适配"),
		}, nil
	}
	return SignInResult{
		Code:    SignInCodeFailure,
		Message: fmt.Sprintf("签到失败，状态码：%d", statusCode),
	}, nil
}

func (c *npClient) GetDownloadUrl(torrent SearchTorrent) (string, error) {
	return torrent.Enclosure, nil
}

func (c *npClient) Details(id string) (TorrentDetail, error) {
	env := map[string]string{"id": id}
	torrentDetail := TorrentDetail{}
	err := data(requestSiteParams{
		site:  c.site,
		reqId: requestIdDetails,
		env:   env,
	}, &torrentDetail, nil)
	if err != nil {
		return torrentDetail, newError(c.site, err, "获取详情异常")
	}
	return torrentDetail, nil
}
