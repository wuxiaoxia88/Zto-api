package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	fmt.Println("调试一键登录流程...")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true), // 在服务器环境下必须使用 headless
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.WindowSize(1280, 800),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	var cookies []*network.Cookie
	var screenshot []byte

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.zt-express.com"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("页面导航中...")
			return nil
		}),
		chromedp.Sleep(10*time.Second),

		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("尝试点击按钮...")
			var result string
			err := chromedp.Evaluate(`
				(function() {
					const btn = Array.from(document.querySelectorAll('button, a, div, span')).find(el => el.textContent.trim() === '一键登录' && el.offsetWidth > 0);
					if (btn) {
						btn.click();
						return 'Clicked: ' + btn.tagName;
					}
					return 'Not Found';
				})()
			`, &result).Do(ctx)
			fmt.Println("点击结果:", result)
			return err
		}),

		chromedp.Sleep(10*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("等待加载中...")
			return nil
		}),

		chromedp.CaptureScreenshot(&screenshot),
		chromedp.ActionFunc(func(ctx context.Context) error {
			os.WriteFile("login_debug.png", screenshot, 0644)
			fmt.Println("截图已保存到 login_debug.png")

			var err error
			cookies, err = network.GetCookies().Do(ctx)
			return err
		}),
	)

	if err != nil {
		log.Println("发生错误:", err)
	}

	fmt.Printf("\nCookies 总数: %d\n", len(cookies))
	for _, c := range cookies {
		fmt.Printf("  %s (%s)\n", c.Name, c.Domain)
	}
}
