package btsite

import (
	"encoding/xml"
)

// 内部使用
type (
	// Torrent 种子
	torrent struct {
		ID                   string   `mapstructure:"id,omitempty"`
		Category             string   `mapstructure:"category,omitempty"`
		Title                string   `mapstructure:"title,omitempty"`
		Details              string   `mapstructure:"details,omitempty"`
		Download             string   `mapstructure:"download,omitempty"`
		Size                 int64    `mapstructure:"size,omitempty"`
		Grabs                int      `mapstructure:"grabs,omitempty"`
		Seeders              int      `mapstructure:"seeders,omitempty"`
		Leechers             int      `mapstructure:"leechers,omitempty"`
		DateElapsed          string   `mapstructure:"date_elapsed,omitempty"`
		DateAdded            string   `mapstructure:"date_added,omitempty"`
		DownloadVolumeFactor float64  `mapstructure:"downloadvolumefactor,omitempty"`
		UploadVolumeFactor   float64  `mapstructure:"uploadvolumefactor,omitempty"`
		Description          string   `mapstructure:"description,omitempty"`
		Labels               []string `mapstructure:"labels,omitempty"`
		HrDays               int      `mapstructure:"hr_days,omitempty"`
	}
	// seeding 做种信息
	seeding struct {
		Size int64 `mapstructure:"size,omitempty"` // 体积，单位字节
	}
	rssResult struct {
		XMLName xml.Name `xml:"rss"`
		Items   []struct {
			Title       string `xml:"title"`
			Description string `xml:"description"`
			Link        string `xml:"link"`
			Enclosure   struct {
				URL    string `xml:"url,attr"`
				Length string `xml:"length,attr"`
			} `xml:"enclosure"`
			Guid    string `xml:"guid"`
			PubDate string `xml:"pubDate"`
		} `xml:"channel>item"`
	}
	signInResult struct {
		SignedIn bool `mapstructure:"signed_in,omitempty"` // 是否签到成功
	}
	markAsReadResult struct {
		Success bool   `mapstructure:"success,omitempty"`
		Message string `mapstructure:"message,omitempty"`
	}
)

// 接口出参相关
type (
	// UserBasicInfo 用户基础信息
	UserBasicInfo struct {
		IsLogin            bool    `mapstructure:"is_login,omitempty"`             // 是否已登录
		SignedIn           bool    `mapstructure:"signed_in,omitempty"`            // 是否已签到
		ID                 string  `mapstructure:"id,omitempty"`                   // 用户 id
		Name               string  `mapstructure:"name,omitempty"`                 // 用户名
		UnreadMessageCount int     `mapstructure:"unread_message_count,omitempty"` // 未读消息数量
		Ratio              float64 `mapstructure:"ratio,omitempty"`                // 分享率
		Uploaded           int64   `mapstructure:"uploaded,omitempty"`             // 上传量，单位字节
		Downloaded         int64   `mapstructure:"downloaded,omitempty"`           // 下载量，单位字节
		Bonus              float64 `mapstructure:"bonus,omitempty"`                // 魔力值

		// 以下为魔力值的辅助字段，外部不要使用。
		Gold   float64 `mapstructure:"Gold,omitempty"`   // 金币
		Silver float64 `mapstructure:"Silver,omitempty"` // 银币
		Copper float64 `mapstructure:"Copper,omitempty"` // 铜币
	}
	// UserDetails 用户详情
	UserDetails struct {
		Level        string `mapstructure:"level,omitempty"`         // 用户等级
		JoinAt       int64  `mapstructure:"join_at,omitempty"`       // 注册时间 时间戳
		LastAccessed int64  `mapstructure:"last_accessed,omitempty"` // 最近访问时间 时间戳
	}
	// SearchTorrent 搜索种子
	SearchTorrent struct {
		ID                   string   // ID
		Category             string   // 分类
		Title                string   // 标题
		Description          string   // 描述
		PageURL              string   // 详情页
		Enclosure            string   // 下载链接
		Grabs                int      // 完成数
		Seeders              int      // 做种人数
		Leechers             int      // 下载人数
		Size                 int64    // 体积
		DownloadVolumeFactor float64  // 下载系数
		UploadVolumeFactor   float64  // 上传系数
		PubDate              string   // 发布时间
		DateElapsed          string   // 剩余时间
		HrDays               int      // HitAndRun 天数
		HitAndRun            bool     // 是否 HitAndRun
		Labels               []string // 标签
	}
	// SeedingStatistics 做种统计
	SeedingStatistics struct {
		Count int   `mapstructure:"count,omitempty"` // 数量
		Size  int64 `mapstructure:"size,omitempty"`  // 体积，单位字节
	}
	// HrTorrent HR 种子
	HrTorrent struct {
		ID                      string `mapstructure:"id,omitempty"`                        // 考核 ID
		Name                    string `mapstructure:"name,omitempty"`                      // 种子名
		Uploaded                string `mapstructure:"uploaded,omitempty"`                  // 上传量
		Downloaded              string `mapstructure:"downloaded,omitempty"`                // 下载量
		ShareRatio              string `mapstructure:"share_ratio,omitempty"`               // 分享率
		DownloadTime            string `mapstructure:"download_time,omitempty"`             // 下载时间，或者统计时间
		NeedSeedTime            string `mapstructure:"need_seed_time,omitempty"`            // 需要做种时间
		RemainingInspectionTime string `mapstructure:"remaining_inspection_time,omitempty"` // 剩余时间
	}
	// Message 未读消息
	Message struct {
		ID      string `mapstructure:"id,omitempty"`      // ID，通常只有 api 方式的站点用的到
		Head    string `mapstructure:"head,omitempty"`    // 标题
		Date    string `mapstructure:"date,omitempty"`    // 时间戳
		Content string `mapstructure:"content,omitempty"` // 内容
		Link    string `mapstructure:"link,omitempty"`    // 详情链接
	}
	// Notice 公告
	Notice struct {
		Title   string `mapstructure:"title,omitempty"`   // 标题
		Date    int64  `mapstructure:"date,omitempty"`    // 时间戳
		Content string `mapstructure:"content,omitempty"` // 内容
	}
	// RssTorrent RSS 拉取的数据
	RssTorrent struct {
		ID          string // ID
		Title       string // 标题
		Enclosure   string // 下载链接
		Size        int64  // 体积，单位字节
		Description string // 描述
		Link        string // 详情页
		PubDate     int64  // 发布时间
	}
	// SignInResult 签到结果
	SignInResult struct {
		Code    SignInCode // 状态码
		Message string     // 提示信息
	}
	// TorrentDetail 种子详情
	TorrentDetail struct {
		Absent     bool `mapstructure:"absent"`     // 种子已不存在
		Free       bool `mapstructure:"free"`       // 是否免费
		DoubleFree bool `mapstructure:"2x_free"`    // 是否双免
		HR         bool `mapstructure:"hr"`         // 是否是 HR 种子
		PeerCount  int  `mapstructure:"peer_count"` // 做种人数
	}
)

// 接口入参相关
type (
	// SearchParams 搜索种子参数
	SearchParams struct {
		Keyword   string // 关键字
		MediaType MediaType
		Page      int
	}
	Site struct {
		Code      string
		Name      string
		UserId    string
		Api       string
		Domain    string
		UserAgent string
		Cookie    string
		Headers   string
		RssUrl    string
	}
	MediaType struct {
		Code string
		Name string
	}
)
