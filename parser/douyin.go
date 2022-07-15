package parser

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-resty/resty/v2"
)

type douYinRes struct {
	StatusCode int `json:"status_code"`
	ItemList   []struct {
		Video struct {
			PlayAddr struct {
				UrlList []string `json:"url_list"`
				Uri     string   `json:"uri"`
			} `json:"play_addr"`
		} `json:"video"`
		Music struct {
			Title   string `json:"title"`
			PlayUrl struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"play_url"`
		} `json:"music"`
		Desc string `json:"desc"`
	} `json:"item_list"`
}

type douYin struct{}

func (d douYin) parseVideoID(videoId string) (*VideoParseInfo, error) {
	parseList, err := d.multiParseVideoID([]string{videoId})
	if err != nil {
		return nil, err
	}
	if len(parseList) <= 0 {
		return nil, errors.New("has no parse info")
	}

	return parseList[0], nil
}

func (d douYin) parseShareUrl(shareUrl string) (*VideoParseInfo, error) {
	client := resty.New()
	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	res, _ := client.R().EnableTrace().Get(shareUrl)
	// 这里会返回err, auto redirect is disabled

	locationRes, err := res.RawResponse.Location()
	if err != nil {
		return nil, err
	}

	videoId := strings.ReplaceAll(strings.Trim(locationRes.Path, "/"), "share/video/", "")
	if len(videoId) <= 0 {
		return nil, errors.New("parse video id from share url fail")
	}

	return d.parseVideoID(videoId)
}

func (d douYin) multiParseVideoID(videoIds []string) ([]*VideoParseInfo, error) {
	// 支持多个videoId批量获取, 用逗号隔开
	itemIds := strings.Join(videoIds, ",")
	reqUrl := "https://www.iesdouyin.com/web/api/v2/aweme/iteminfo/?item_ids=" + itemIds
	client := resty.New()
	res, err := client.R().Get(reqUrl)
	if err != nil {
		return nil, err
	}

	douYinRes := &douYinRes{}
	json.Unmarshal(res.Body(), douYinRes)

	parseList := make([]*VideoParseInfo, 0, len(douYinRes.ItemList))
	for _, item := range douYinRes.ItemList {
		if len(item.Video.PlayAddr.UrlList) <= 0 {
			continue
		}
		videoPlayAddr := strings.ReplaceAll(item.Video.PlayAddr.UrlList[0], "/playwm/", "/play/")
		parseItem := &VideoParseInfo{
			Title:    item.Desc,
			VideoUrl: videoPlayAddr,
			MusicUrl: item.Music.PlayUrl.Uri,
		}
		parseList = append(parseList, parseItem)
	}

	return parseList, nil
}