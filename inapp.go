package go_inapp_parser

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/valyala/fasthttp"
)

const GoogleURL = "https://play.google.com/store/apps/details?id="
const AppleURL = "https://amp-api.apps.apple.com/v1/catalog/RU/apps/"

var ErrStatus = fmt.Errorf("bad response status")
var AppleIconReplacer = strings.NewReplacer("{f}", "png", "{w}", "460", "{h}", "0", "{c}", "w")

var defSetting Setting = Setting{
	Timeout:   100 * time.Millisecond,
	AppleKey:  "",
	GoogleKey: "",
}

type Setting struct {
	Timeout   time.Duration
	AppleKey  string
	GoogleKey string
}

type Parser struct {
	setting Setting
}

type Option func(p *Parser)

func SetTimout(v time.Duration) Option {
	return func(p *Parser) {
		p.setting.Timeout = v
	}
}

func SetAppleApiKey(v string) Option {
	return func(p *Parser) {
		p.setting.AppleKey = v
	}
}

func New(opts ...Option) *Parser {
	proto := new(Parser)
	proto.setting = defSetting
	for _, opt := range opts {
		opt(proto)
	}
	return proto
}

// ParseByBundleID - parsing by bundle_id
func (p *Parser) ParseByBundleID(id string) (*Info, error) {
	if len(strings.Split(id, ".")) > 1 {
		return ParseGoogle(id, p.setting.Timeout)
	} else {
		return ParseApple(id, p.setting.AppleKey, p.setting.Timeout)
	}
}

// ParseGoogle - parsing google app
func ParseGoogle(id string, timeout time.Duration) (*Info, error) {

	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(res)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(GoogleURL + id)
	req.Header.SetMethod(fasthttp.MethodGet)

	if err := doFollowRedirectsTimeout(req, res, timeout); err != nil {
		return nil, err
	} else if res.StatusCode() != fasthttp.StatusOK {
		return nil, ErrStatus
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(res.Body()))
	if err != nil {
		return nil, err
	}

	model := &Info{
		ID:   id,
		Url:  GoogleURL + id,
		Type: Android,
	}

	// load category
	doc.Find("a[itemprop=\"genre\"]").Each(func(index int, tHtml *goquery.Selection) {
		model.Category = strings.Split(strings.ToLower(tHtml.Text()), ",")
	})

	// load name
	doc.Find("h1[itemprop=\"name\"]>span").Each(func(index int, tHtml *goquery.Selection) {
		model.Name = tHtml.Text()
	})

	// load publisher
	doc.Find("a[href*=\"/store/apps/dev\"]").Each(func(index int, tHtml *goquery.Selection) {
		model.Publisher = tHtml.Text()
	})

	// load icon
	doc.Find("img[alt=\"Cover art\"]").Each(func(index int, tHtml *goquery.Selection) {
		if val, ext := tHtml.Attr("src"); ext {
			model.Icon = val
		}
	})

	// load rate
	doc.Find("div>div>c-wiz>div>div>div>div>div>c-wiz>div>c-wiz>div>div").Each(func(index int, tHtml *goquery.Selection) {
		if index == 0 {
			model.Rate, _ = strconv.ParseFloat(tHtml.Text(), 64)
		}
	})

	return model, nil
}

// ParseApple - parsing apple app
func ParseApple(id string, key string, timeout time.Duration) (*Info, error) {

	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(res)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(AppleURL + id + "?platform=web&additionalPlatforms=appletv%2Cipad%2Ciphone%2Cmac&extend=description%2CdeveloperInfo%2CeditorialVideo%2Ceula%2CfileSizeByDevice%2CmessagesScreenshots%2CprivacyPolicyUrl%2CprivacyPolicyText%2CpromotionalText%2CscreenshotsByType%2CsupportURLForLanguage%2CversionHistory%2CvideoPreviewsByType%2CwebsiteUrl&include=genres%2Cdeveloper%2Creviews%2Cmerchandised-in-apps%2Ccustomers-also-bought-apps%2Cdeveloper-other-apps%2Capp-bundles%2Ctop-in-apps%2Ceula&l=en-us")
	req.Header.Set(fasthttp.HeaderContentType, "application/json")
	req.Header.Set(fasthttp.HeaderAuthorization, "Bearer "+key)
	req.Header.SetMethod(fasthttp.MethodGet)

	if err := doFollowRedirectsTimeout(req, res, timeout); err != nil {
		return nil, err
	} else if res.StatusCode() != fasthttp.StatusOK {
		return nil, ErrStatus
	}

	data, err := newJSON(res.Body())
	if err != nil {
		return nil, err
	}

	model := &Info{
		ID:   id,
		Type: Apple,
	}

	// parse category
	for _, v := range data.GetPath("data").GetIndex(0).GetPath("relationships", "genres", "data").MustArray() {
		if v1, ok := v.(map[string]interface{})["attributes"]; ok {
			if v2, ok := v1.(map[string]interface{})["name"]; ok {
				if v3, ok := v2.(string); ok {
					model.Category = append(model.Category, strings.ToLower(v3))
				}
			}
		}
	}

	// parse icon
	icon := data.GetPath("data").GetIndex(0).GetPath("attributes", "platformAttributes", "ios", "artwork", "url").MustString()
	model.Icon = AppleIconReplacer.Replace(icon)

	// parse other
	model.Name = data.GetPath("data").GetIndex(0).GetPath("attributes", "name").MustString()
	model.Publisher = data.GetPath("data").GetIndex(0).GetPath("attributes", "artistName").MustString()
	model.Url = data.GetPath("data").GetIndex(0).GetPath("attributes", "url").MustString()
	model.Rate = data.GetPath("data").GetIndex(0).GetPath("attributes", "userRating", "value").MustFloat64()

	return model, nil
}
