package btsite

import (
	"github.com/heibizi/go-siteadapt"
	"strings"
	"time"
)

type (
	// mtClient 馒头客户端
	mtClient struct {
		site *Site
		*npClient
	}
	mtMyPeerStatus struct {
		Leecher int `mapstructure:"leecher"`
		Seeder  int `mapstructure:"seeder"`
	}
	mtMsgNotifyStatistic struct {
		Count  int `mapstructure:"count"`
		UnMake int `mapstructure:"un_make"`
	}
	mtProfile struct {
		CreatedDate      int64   `mapstructure:"created_date,omitempty"`
		LastModifiedDate int64   `mapstructure:"last_modified_date,omitempty"`
		Username         string  `mapstructure:"username,omitempty"`
		Uploaded         int64   `mapstructure:"uploaded,omitempty"`
		Downloaded       int64   `mapstructure:"downloaded,omitempty"`
		ShareRate        float64 `mapstructure:"share_rate,omitempty"`
		Bonus            float64 `mapstructure:"bonus,omitempty"`
		Role             string  `mapstructure:"role,omitempty"`
	}
	mtUserTorrent struct {
		Size int64 `mapstructure:"size"`
	}
	mtSysRole struct {
		Id      string `mapstructure:"id"`
		NameChs string `mapstructure:"name_chs"`
		NameEng string `mapstructure:"name_eng"`
	}
)

const (
	requestIdMTMyPeerStatus       requestId = "my_peer_status"
	requestIdMTMsgNotifyStatistic requestId = "msg_notify_statistic"
	requestIdMTProfile            requestId = "profile"
	requestIdMTUserTorrentList    requestId = "user_torrent_list"
	requestIdMTSysRoleList        requestId = "sys_role_list"
	requestIdMTGenDLToken         requestId = "gen_dl_token"
)

func (c *mtClient) UserBasicInfo() (UserBasicInfo, error) {
	mp, err := c.memberProfile()
	if err != nil {
		return UserBasicInfo{}, err
	}
	mns, err := c.msgNotifyStatistic()
	if err != nil {
		return UserBasicInfo{}, err
	}
	return UserBasicInfo{
		IsLogin:            len(mp.Username) > 0,
		ID:                 c.site.UserId,
		Name:               mp.Username,
		UnreadMessageCount: mns.UnMake,
		Ratio:              mp.ShareRate,
		Uploaded:           mp.Uploaded,
		Downloaded:         mp.Downloaded,
		Bonus:              mp.Bonus,
	}, nil
}

func (c *mtClient) UserDetails() (UserDetails, error) {
	mp, err := c.memberProfile()
	if err != nil {
		return UserDetails{}, err
	}
	level := ""
	srl, err := c.sysRoleList()
	if err != nil {
		return UserDetails{}, err
	}
	if srl != nil {
		for _, role := range srl {
			if mp.Role == role.Id {
				level = role.NameChs + " " + role.NameEng
			}
		}
	}
	return UserDetails{
		Level:        level,
		JoinAt:       mp.CreatedDate,
		LastAccessed: mp.LastModifiedDate,
	}, nil
}

func (c *mtClient) Search(searchParams SearchParams) ([]SearchTorrent, error) {
	mode := "movie"
	if searchParams.MediaType == Tv {
		mode = "tvshow"
	} else if searchParams.MediaType == Anime {
		mode = "normal"
	}
	body := map[string]any{
		"mode":    mode,
		"visible": 1,
		// 馒头从 1 开始
		"pageNumber": searchParams.Page + 1,
		"pageSize":   100,
	}
	if len(searchParams.Keyword) > 0 {
		body["keyword"] = searchParams.Keyword
	}
	site := c.site
	var torrents []torrent
	domain := ""
	err := list(requestSiteParams{
		site:  site,
		reqId: requestIdSearch,
		body:  body,
	}, &torrents, func(result siteadapt.ListResult) {
		domain = result.Domain
	})
	if err != nil {
		return nil, newError(site, err, "搜索异常")
	}
	var searchTorrents []SearchTorrent
	for _, torrent := range torrents {
		pageUrl, err := JoinURL(domain, torrent.Details)
		if err != nil {
			return nil, err
		}
		var labels []string
		for _, label := range torrent.Labels {
			for _, s := range strings.Split(label, "|") {
				labels = append(labels, s)
			}
		}
		search := SearchTorrent{
			ID:                   torrent.ID,
			Category:             torrent.Category,
			Title:                torrent.Title,
			Description:          torrent.Description,
			PageURL:              pageUrl,
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
			Labels:               labels,
		}
		searchTorrents = append(searchTorrents, search)
	}
	return searchTorrents, nil
}

func (c *mtClient) SeedingStatistics() (SeedingStatistics, error) {
	seeding := SeedingStatistics{}
	for pageNumber := 1; ; pageNumber++ {
		tl, err := c.userTorrentList(pageNumber)
		if err != nil {
			return seeding, err
		}
		if tl == nil {
			break
		}
		for _, ut := range tl {
			seeding.Count = seeding.Count + 1
			seeding.Size = seeding.Size + ut.Size
		}
		time.Sleep(500 * time.Millisecond)
	}
	return seeding, nil
}

func (c *mtClient) MyHr() ([]HrTorrent, error) {
	// 无 hr
	return nil, nil
}

func (c *mtClient) SignIn() (SignInResult, error) {
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
	return SignInResult{
		Code:    SignInCodeSuccess,
		Message: "模拟登录成功",
	}, nil
}

func (c *mtClient) UnreadMessages(detail bool) ([]Message, error) {
	var o []Message
	fromDataEnv := map[string]string{"pageNumber": "1"}
	err := list(requestSiteParams{
		site:  c.site,
		reqId: requestIdUnreadMessages,
		env:   fromDataEnv,
	}, &o, nil)
	if err != nil {
		return nil, newError(c.site, err, "获取未读消息异常")
	}
	if detail {
		var ids []string
		for _, message := range o {
			ids = append(ids, message.ID)
		}
		err = c.markAsRead(ids)
		if err != nil {
		}
		return nil, err
	}
	return o, nil
}

func (c *mtClient) myPeerStatus() (mtMyPeerStatus, error) {
	o := mtMyPeerStatus{}
	err := data(requestSiteParams{
		site:  c.site,
		reqId: requestIdMTMyPeerStatus,
	}, &o, nil)
	if err != nil {
		return o, newError(c.site, err, "做种信息异常")
	}
	return o, nil
}

func (c *mtClient) msgNotifyStatistic() (mtMsgNotifyStatistic, error) {
	o := mtMsgNotifyStatistic{}
	err := data(requestSiteParams{
		site:  c.site,
		reqId: requestIdMTMsgNotifyStatistic,
	}, &o, nil)
	if err != nil {
		return o, newError(c.site, err, "消息通知统计异常")
	}
	return o, nil
}

// memberProfile 获取个人资料
func (c *mtClient) memberProfile() (mtProfile, error) {
	o := mtProfile{}
	err := data(requestSiteParams{
		site:  c.site,
		reqId: requestIdMTProfile,
	}, &o, nil)
	if err != nil {
		return o, newError(c.site, err, "个人资料异常")
	}
	return o, nil
}

func (c *mtClient) userTorrentList(pageNumber int) ([]mtUserTorrent, error) {
	var o []mtUserTorrent
	var body = make(map[string]any)
	body["userid"] = c.site.UserId
	body["type"] = "SEEDING"
	body["pageNumber"] = pageNumber
	body["pageSize"] = 100
	err := list(requestSiteParams{
		site:  c.site,
		reqId: requestIdMTUserTorrentList,
		body:  body,
	}, &o, nil)
	if err != nil {
		return nil, newError(c.site, err, "用户做种列表异常")
	}
	return o, nil
}

func (c *mtClient) sysRoleList() ([]mtSysRole, error) {
	var o []mtSysRole
	err := list(requestSiteParams{
		site:  c.site,
		reqId: requestIdMTSysRoleList,
	}, &o, nil)
	if err != nil {
		return nil, newError(c.site, err, "角色列表异常")
	}
	return o, nil
}

func (c *mtClient) genDlToken(torrent SearchTorrent) (string, error) {
	formDataEnv := map[string]string{"id": torrent.ID}
	m := make(map[string]any)
	err := data(requestSiteParams{
		site:  c.site,
		reqId: requestIdMTGenDLToken,
		env:   formDataEnv,
	}, &m, nil)
	if err != nil {
		return "", err
	}
	return m["url"].(string), nil
}

// markAsRead 未读消息设为已读
func (c *mtClient) markAsRead(ids []string) error {
	r := markAsReadResult{}
	err := data(requestSiteParams{
		site:  c.site,
		reqId: requestIdMarkAsRead,
		env:   map[string]string{"ids": strings.Join(ids, ",")},
	}, &r, nil)
	if err != nil {
		return newError(c.site, err, "未读消息设为已读异常")
	}
	if r.Success {
		return nil
	}
	return newError(c.site, err, "未读消息设为已读失败: %s", r.Message)
}
