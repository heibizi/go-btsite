package btsite

import (
	"fmt"
	"os"
	"path/filepath"
)

var SiteHelper = new(helper)

type helper struct {
}

func (sh *helper) GetFaviconPath(site Site, storePath string) (string, error) {
	// 读取 data/favicon
	filename := filepath.Join(storePath, site.Code+".ico")
	_, err := os.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		return "", fmt.Errorf("站点 favicon 不存在: %s", filename)
	}
	return filename, nil
}

func (sh *helper) SaveFavicon(data []byte, site *Site, storePath string) error {
	// 保存到 data/favicon
	filename := filepath.Join(storePath, site.Code+".ico")
	// 确保目录存在
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %s", dir)
	}
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建 favicon 失败: %s", filename)
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("保存 favicon 失败: %s", filename)
	}
	return nil
}

func (sh *helper) GetConfigByCode(code string) (Config, error) {
	// 根据 code 获取站点配置
	for _, siteConfig := range globalConfig.Configs {
		if siteConfig.ID == code {
			return siteConfig, nil
		}
	}
	return Config{}, fmt.Errorf("站点配置不存在: %s", code)
}

func (sh *helper) GetDomain(site Site) (string, error) {
	if len(site.Domain) > 0 {
		return site.Domain, nil
	}
	sc, err := sh.GetConfigByCode(site.Code)
	if err != nil {
		return "", err
	}
	return sc.Domain, nil
}

func (sh *helper) GetApi(site Site) (string, error) {
	if len(site.Api) > 0 {
		return site.Api, nil
	}
	sc, err := sh.GetConfigByCode(site.Code)
	if err != nil {
		return "", err
	}
	return sc.Api, nil
}

func (sh *helper) AllSupportedSites() []Config {
	return globalConfig.Configs
}
