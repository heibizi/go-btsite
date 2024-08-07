package btsite

import (
	"encoding/json"
	"fmt"
	"github.com/heibizi/go-siteadapt"
	"os"
	"path/filepath"
)

var globalConfig AdaptCfg

type (
	// AdaptCfg 站点适配配置
	AdaptCfg struct {
		Common  map[string]Config
		Configs []Config
	}
	// AdaptMediaCat 媒体分类
	AdaptMediaCat struct {
		ID      string `mapstructure:"id"`
		Cat     string `mapstructure:"cat"`
		Desc    string `mapstructure:"desc"`
		Default bool   `mapstructure:"default"`
	}
	Config struct {
		siteadapt.Config
		Public       bool   `mapstructure:"public"`        // 是否为公开站点
		Schema       string `mapstructure:"schema"`        // 系统架构，对应 siteSchema
		ReuseSchema  string `mapstructure:"reuse_schema"`  // 当前系统架构没有公共配置时，使用该系统架构的公共配置
		CountMessage bool   `mapstructure:"count_message"` // 未读消息从未读消息列表统计数量
		// 必需参数
		Required struct {
			UserID bool `mapstructure:"user_id"`
			SignIn bool `mapstructure:"sign_in"`
			Cookie bool `mapstructure:"cookie"`
		} `mapstructure:"required"`
		Categories struct {
			Movie     []AdaptMediaCat `mapstructure:"movie"`
			TV        []AdaptMediaCat `mapstructure:"tv"`
			Field     string          `mapstructure:"field"`
			Delimiter string          `mapstructure:"delimiter"`
		} `mapstructure:"category"`
		// Price 促销配置
		Price struct {
			HasFree   bool `mapstructure:"has_free"`    // 是否有 FREE
			Has2XFree bool `mapstructure:"has_2x_free"` // 是否有 2XFree
			HasHR     bool `mapstructure:"has_hr"`      // 是否有 HR
		} `mapstructure:"price"`
	}
)

// InitConfig 初始化站点适配文件
func InitConfig(path string) {
	if len(path) == 0 {
		panic("站点配置文件路径未配置")
	}
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		panic(fmt.Errorf("站点配置文件路径错误: %s", path))
	}
	conf := AdaptCfg{}
	common, err := loadCommonConfigs(path)
	if err != nil {
		panic(err)
	}
	conf.Common = common
	configs, err := loadSiteConfigs(path)
	if err != nil {
		panic(err)
	}
	conf.Configs = configs
	for i := range conf.Configs {
		sc := &conf.Configs[i]
		// schema
		schemaSc, exists := conf.Common[sc.Schema]
		if !exists {
			// reuse schema
			schemaSc, exists = conf.Common[sc.ReuseSchema]
		}
		if exists {
			for schemaRdName, schemaRd := range schemaSc.RequestDefinitions {
				rd, exists := sc.RequestDefinitions[schemaRdName]
				if exists {
					extend(&rd, &schemaRd)
					sc.RequestDefinitions[schemaRdName] = rd
				} else {
					sc.RequestDefinitions[schemaRdName] = schemaRd
				}
			}
		}
	}
	globalConfig = conf
}

func extend(rd *siteadapt.RequestDefinition, schemaRd *siteadapt.RequestDefinition) {
	if rd.Parser == "" {
		rd.Parser = schemaRd.Parser
	}
	if rd.Method == "" {
		rd.Method = schemaRd.Method
	}
	if !rd.DisabledExtends.Path && rd.Path == "" {
		rd.Path = schemaRd.Path
	}
	if !rd.DisabledExtends.List && rd.List == nil {
		rd.List = schemaRd.List
	}
	if !rd.DisabledExtends.Headers && rd.Headers == nil {
		rd.Headers = schemaRd.Headers
	}
	if !rd.DisabledExtends.RequiredHeaders && rd.RequiredHeaders == nil {
		rd.RequiredHeaders = schemaRd.RequiredHeaders
	}
	if !rd.DisabledExtends.Params && rd.Params == nil {
		rd.Params = schemaRd.Params
	}
	if !rd.DisabledExtends.FormData && rd.FormData == nil {
		rd.FormData = schemaRd.FormData
	}
	if !rd.DisabledExtends.SuccessStatusCodes && rd.SuccessStatusCodes == nil {
		rd.SuccessStatusCodes = schemaRd.SuccessStatusCodes
	}
	if !rd.DisabledExtends.List && rd.List == nil {
		rd.List = schemaRd.List
	}
	if !rd.DisabledExtends.Fields && rd.Fields == nil {
		rd.Fields = schemaRd.Fields
	}
	if !rd.DisabledExtends.Field {
		// 继承字段
		for schemaRdFieldName, schemaRdField := range schemaRd.Fields {
			if _, exists := rd.Fields[schemaRdFieldName]; !exists {
				rd.Fields[schemaRdFieldName] = schemaRdField
			}
		}
	}
}

// listFiles 函数递归获取指定目录下的所有文件内容，并返回一个二维字节切片
func listFiles(dir string) ([][]byte, error) {
	var filesContent [][]byte
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			filesContent = append(filesContent, content)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return filesContent, nil
}

func readConfig[T any](file []byte, output *T) error {
	var data interface{}
	err := json.Unmarshal(file, &data)
	if err != nil {
		return fmt.Errorf("解析配置异常: %v", err)
	}
	if err := siteadapt.WeakDecode(data, output); err != nil {
		return fmt.Errorf("解析配置异常: %v", err)
	}
	return nil
}

// loadCommonConfigs 加载公共配置
func loadCommonConfigs(path string) (map[string]Config, error) {
	commons := make(map[string]Config)
	path = filepath.Join(path, "commons")
	files, err := listFiles(path)
	if err != nil {
		return nil, fmt.Errorf("加载公共配置异常: %v", err)
	}
	for _, file := range files {
		var common Config
		configReader := siteadapt.NewConfigReader(file)
		config, err := configReader.Read()
		if err != nil {
			return nil, fmt.Errorf("加载公共配置异常: %v", err)
		}
		common.Config = *config
		checkSearchFields(common.RequestDefinitions)
		commons[common.ID] = common
	}
	return commons, nil
}

// loadSiteConfigs 加载站点配置
func loadSiteConfigs(path string) ([]Config, error) {
	var siteConfigs []Config
	path = filepath.Join(path, "sites")
	files, err := listFiles(path)
	if err != nil {
		return nil, fmt.Errorf("加载站点配置异常: %v", err)
	}
	for _, file := range files {
		var sc Config
		err := readConfig(file, &sc)
		if err != nil {
			return nil, err
		}
		configReader := siteadapt.NewConfigReader(file)
		config, err := configReader.Read()
		if err != nil {
			return nil, fmt.Errorf("加载站点配置异常: %v", err)
		}
		sc.Config = *config
		checkSearchFields(sc.RequestDefinitions)
		siteConfigs = append(siteConfigs, sc)
	}
	return siteConfigs, err
}

var validFields = []string{"id", "category", "title", "details", "download", "size", "grabs", "seeders",
	"leechers", "date_elapsed", "date_added", "downloadvolumefactor", "uploadvolumefactor", "description",
	"labels", "hr_days", "imdbid"}

// checkSearchFields 检查搜索字段，为了 json 的简洁性强制性要求不能乱配置
func checkSearchFields(rds map[string]siteadapt.RequestDefinition) {
	fields := rds["search"].Fields
	for name := range fields {
		valid := false
		for _, validField := range validFields {
			if name == validField {
				valid = true
				break
			}
		}
		if !valid {
			panic(fmt.Errorf("为了适配文件的简洁性，强制要求 field 按照规则配置，搜索字段不合法: %s", name))
		}
	}
}
