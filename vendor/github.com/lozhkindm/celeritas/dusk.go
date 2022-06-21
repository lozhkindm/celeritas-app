package celeritas

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
)

func (c *Celeritas) TakeScreenshot(url, testname string, wight, height float64) error {
	img, err := c.GetPage(url).Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
		Clip: &proto.PageViewport{
			X:      0,
			Y:      0,
			Width:  wight,
			Height: height,
			Scale:  1,
		},
		FromSurface: true,
	})
	if err != nil {
		return err
	}

	file := fmt.Sprintf(
		"%s/screenshots/%s-%s.png",
		c.RootPath,
		testname,
		time.Now().Format("2006-01-02-15-04-05.000000"),
	)
	if err := utils.OutputFile(file, img); err != nil {
		return err
	}
	return nil
}

func (c *Celeritas) GetPage(url string) *rod.Page {
	return rod.New().
		MustConnect().
		MustIgnoreCertErrors(true).
		MustPage(url).
		MustWaitLoad()
}

func (c *Celeritas) GetElementByID(page *rod.Page, ID string) *rod.Element {
	return page.MustElement("#" + ID)
}
