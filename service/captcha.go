package service

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/dchest/captcha"
)

var Captcha *CaptchaWrap

func init() {
	Captcha = &CaptchaWrap{
		digitLength: 4,
		expire:      5 * time.Minute,
		width:       captcha.StdWidth,
		height:      captcha.StdHeight,
	}
}

type CaptchaWrap struct {
	digitLength int
	expire      time.Duration
	width       int
	height      int
}

type CaptchaValue struct {
	Value    string
	CreateAt time.Time
}

func (c *CaptchaWrap) NewImage() (randValue CaptchaValue, image *captcha.Image) {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	id := strings.TrimRight(base64.URLEncoding.EncodeToString(buf), "=")
	digits := captcha.RandomDigits(c.digitLength)
	image = captcha.NewImage(id, digits, c.width, c.height)
	// 生成图片是按照 []byte{1,0,2,4}（ascii字符数值） => image 1024 ，变成字符串"1024"需要每个byte + '0'
	for i := range digits {
		digits[i] += '0'
	}
	return CaptchaValue{
		Value:    string(digits),
		CreateAt: time.Now(),
	}, image
}

func (c *CaptchaWrap) NewImageToSession(saveKey string, ctx *context.Context) error {
	capValue, image := c.NewImage()
	if err := ctx.Input.CruSession.Set(saveKey, capValue); err != nil {
		return err
	} else if _, err = image.WriteTo(ctx.ResponseWriter); err != nil {
		return err
	}
	return nil
}

func (c *CaptchaWrap) Verify(saveKey string, value string, ctx *context.Context) bool {
	capValue, ok := ctx.Input.CruSession.Get(saveKey).(CaptchaValue)
	if ok && value == capValue.Value && time.Now().Before(capValue.CreateAt.Add(c.expire)) {
		if err := ctx.Input.CruSession.Delete(saveKey); err != nil {
			logs.Error(err)
		}
		return true
	}
	return false
}
