package iron

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	RunMode            string `json:"RunMode"`
	ServeType          string `json:"ServeType"`
	ServeStr           string `json:"ServeStr"`
	ListenStr          string `json:"ListenStr"`
	LogPath            string `json:"LogPath"`
	AccessWhiteListStr string `json:"AccessWhiteList"`

	SiteViewDir              string `json:"SiteViewDir"`
	SiteStaticBasePath       string `json:"SiteStaticBasePath"`
	SiteStaticUploadBasePath string `json:"SiteStaticUploadBasePath"`

	BaseDir           string   `json:"-"`
	AccessWhiteList   []string `json:"-"`
	Log               *os.File `json:"-"`
	IsTMPLAutoRefresh bool     `json:"-"`

	HttpsListenStr string `json:"HttpsListenStr"`
	HttpsCertPath  string `json:"HttpsCertPath"`
	HttpsKeyPath   string `json:"HttpsKeyPath"`
}

func (p *Server) loadOptions(options Options) error {
	var err error
	if err = p.sanitizeOptions(&options); err != nil {
		return err
	}
	p.Options = options
	return nil
}

func (p *Server) sanitizeOptions(options *Options) error {
	var err error

	if options.BaseDir, err = os.Getwd(); err != nil {
		return err
	}

	if options.SiteViewDir == "" {
		options.SiteViewDir = filepath.Join(options.BaseDir, "view")
	}

	switch options.RunMode {
	case "dev", "test", "proc":
		break
	default:
		options.RunMode = "dev"
	}
	switch options.RunMode {
	case "dev":
		options.IsTMPLAutoRefresh = true
	case "test":
		options.IsTMPLAutoRefresh = true
	case "proc":
		options.IsTMPLAutoRefresh = false
	}

	switch options.ServeType {
	case "fcgi", "server":
		break
	default:
		options.ServeType = "server"
	}

	if options.LogPath != "" {
		if nil != options.Log {
			options.Log.Close()
		}
		options.Log, err = os.OpenFile(options.LogPath, os.O_CREATE|os.O_APPEND|os.O_RDWR|os.O_SYNC, 0755)
		if err != nil {
			return err
		}
		log.SetOutput(options.Log)
	}

	options.AccessWhiteList = nil
	if options.AccessWhiteListStr != "" {
		options.AccessWhiteList = strings.Split(options.AccessWhiteListStr, ",")
		for i := range options.AccessWhiteList {
			options.AccessWhiteList[i] = strings.TrimSpace(options.AccessWhiteList[i])
		}
	}

	return nil
}

func LoadOptionsFile(optionsFilePath string, options interface{}) error {
	var (
		err     error
		content []byte
	)

	content, err = ioutil.ReadFile(optionsFilePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, options)
	if err != nil {
		return err
	}

	return nil
}
