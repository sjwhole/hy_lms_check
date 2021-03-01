package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
	"math"
)

func main() {
	var uid, upw string
	uid = "user"
	upw = "password"

	var buf []byte

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		// Set the headless flag to false to display the browser window
		//chromedp.Flag("headless", false),
		//chromedp.Flag("start-fullscreen", true),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	err := chromedp.Run(ctx,
		Login(uid, upw),
	)
	if err != nil {
		log.Fatal(err)
	}

	var nodes []*cdp.Node

	err = chromedp.Run(ctx,
		chromedp.Nodes("a[class='ic-DashboardCard__link']", &nodes))

	if err != nil {
		_ = fmt.Errorf("could not navigate to page: %v", err)
	}

	var lectureCode []string
	var lectureName, miss string
	for _, n := range nodes {
		info := n.AttributeValue("href")
		info = info[len(info)-5:]
		lectureCode = append(lectureCode, info)
	}

	for _, code := range lectureCode {
		_ = chromedp.Run(ctx, GetAttendanceInfo(&lectureName, &code, &miss))
		//fmt.Println(lectureName, miss)

		//if err := chromedp.Run(ctx, elementScreenshot("//*[@id=\"root\"]", &buf)); err != nil {
		//	log.Fatal(err)
		//}
		//if err := ioutil.WriteFile(lectureName+".png", buf, 0o644); err != nil {
		//	log.Fatal(err)
		//}

		if err := chromedp.Run(ctx, fullScreenshot(90, &buf)); err != nil {
			log.Fatal(err)
		}
		if err := ioutil.WriteFile(lectureName+"_full.png", buf, 0o644); err != nil {
			log.Fatal(err)
		}
	}
}

func Login(id string, password string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate("https://learning.hanyang.ac.kr"),
		chromedp.WaitVisible("uid", chromedp.ByID),
		chromedp.SetValue("uid", id, chromedp.ByID),
		chromedp.SetValue("upw", password, chromedp.ByID),
		chromedp.Click("//*[@id=\"login_btn\"]", chromedp.BySearch),
		chromedp.WaitVisible("//*[@id=\"DashboardCard_Container\"]/div/div[1]/div/a", chromedp.BySearch),
	}
}

func GetAttendanceInfo(name *string, code *string, miss *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate("https://learning.hanyang.ac.kr/courses/" + *code + "/external_tools/9"),
		chromedp.WaitVisible("/html/body/div[2]/div[2]/div[1]/div/nav/ul/li[3]/a/span", chromedp.BySearch),
		chromedp.Text("/html/body/div[2]/div[2]/div[1]/div/nav/ul/li[3]/a/span", name, chromedp.BySearch),
		chromedp.Text("/html/body/div/div/div/div/div[1]/div[2]/span[6]", miss, chromedp.BySearch),
	}
}

//func elementScreenshot(sel string, res *[]byte) chromedp.Tasks {
//	return chromedp.Tasks{
//		chromedp.WaitVisible(sel, chromedp.BySearch),
//		chromedp.Screenshot(sel, res, chromedp.NodeVisible, chromedp.BySearch),
//	}
//}

func fullScreenshot(quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// capture screenshot
			*res, err = page.CaptureScreenshot().
				WithQuality(quality).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
	}
}
