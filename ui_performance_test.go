package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/chromedp/cdproto/performance"
	"github.com/chromedp/chromedp"
)

func TestUIPerformance(t *testing.T) {
	// Create a Chrome instance
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Create a timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Enable performance metrics collection
	perfMetrics := make(map[string]float64)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *performance.Metrics:
			for _, m := range ev.Metrics {
				perfMetrics[m.Name] = m.Value
			}
		}
	})

	var tests = []struct {
		name     string
		path     string
		elements []string
		maxLoad  time.Duration
	}{
		{
			name: "Server List Page",
			path: "/servers",
			elements: []string{
				"table.server-list",
				".server-list tbody tr",
			},
			maxLoad: 3 * time.Second,
		},
		{
			name: "Dashboard Overview",
			path: "/dashboard",
			elements: []string{
				".dashboard-stats",
				".os-distribution-chart",
				".resource-usage-chart",
			},
			maxLoad: 2 * time.Second,
		},
		{
			name: "Server Details",
			path: "/servers/1",
			elements: []string{
				".server-info",
				".service-list",
				".metrics-chart",
			},
			maxLoad: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime := time.Now()

			// Navigate to the page
			url := fmt.Sprintf("http://localhost:3000%s", tt.path)
			var loadTime time.Duration

			err := chromedp.Run(ctx,
				chromedp.Navigate(url),
				chromedp.ActionFunc(func(ctx context.Context) error {
					loadTime = time.Since(startTime)
					return nil
				}),
			)
			if err != nil {
				t.Fatalf("Failed to navigate to %s: %v", url, err)
			}

			// Check page load time
			if loadTime > tt.maxLoad {
				t.Errorf("Page load time exceeded maximum: got %v, want <= %v", loadTime, tt.maxLoad)
			}
			t.Logf("Page load time: %v", loadTime)

			// Verify all required elements are present and rendered
			for _, selector := range tt.elements {
				var visible bool
				err := chromedp.Run(ctx,
					chromedp.WaitVisible(selector),
					chromedp.Evaluate(`document.querySelector("`+selector+`") !== null`, &visible),
				)
				if err != nil {
					t.Errorf("Element %s not found or not visible: %v", selector, err)
				}
			}

			// Collect performance metrics
			var metrics []*performance.Metrics
			err = chromedp.Run(ctx,
				performance.Enable(),
				performance.GetMetrics(&metrics),
			)
			if err != nil {
				t.Fatalf("Failed to collect metrics: %v", err)
			}

			// Log key performance metrics
			t.Log("\nPerformance Metrics:")
			t.Logf("DOM Content Loaded: %.2fms", perfMetrics["DOMContentLoaded"]*1000)
			t.Logf("First Paint: %.2fms", perfMetrics["FirstPaint"]*1000)
			t.Logf("First Contentful Paint: %.2fms", perfMetrics["FirstContentfulPaint"]*1000)
			t.Logf("JS Heap Size: %.2f MB", perfMetrics["JSHeapUsedSize"]/1024/1024)

			// Test interaction performance
			if tt.name == "Server List Page" {
				// Test sorting
				startTime = time.Now()
				err = chromedp.Run(ctx,
					chromedp.Click("th.sortable"),
					chromedp.WaitVisible(".server-list tbody tr"),
				)
				if err != nil {
					t.Errorf("Failed to test sorting: %v", err)
				}
				sortTime := time.Since(startTime)
				t.Logf("Sort operation time: %v", sortTime)

				// Test filtering
				startTime = time.Now()
				err = chromedp.Run(ctx,
					chromedp.SendKeys(".search-input", "Windows"),
					chromedp.WaitVisible(".server-list tbody tr"),
				)
				if err != nil {
					t.Errorf("Failed to test filtering: %v", err)
				}
				filterTime := time.Since(startTime)
				t.Logf("Filter operation time: %v", filterTime)

				// Verify reasonable operation times
				if sortTime > 500*time.Millisecond {
					t.Errorf("Sort operation too slow: %v", sortTime)
				}
				if filterTime > 500*time.Millisecond {
					t.Errorf("Filter operation too slow: %v", filterTime)
				}
			}

			// Test scrolling performance on server list
			if tt.name == "Server List Page" {
				startTime = time.Now()
				err = chromedp.Run(ctx,
					chromedp.Evaluate(`
						window.scrollTo({
							top: document.body.scrollHeight,
							behavior: 'smooth'
						});
					`, nil),
					chromedp.Sleep(1*time.Second),
				)
				if err != nil {
					t.Errorf("Failed to test scrolling: %v", err)
				}
				scrollTime := time.Since(startTime)
				t.Logf("Scroll operation time: %v", scrollTime)

				// Check for frame drops during scroll
				var dropRate float64
				err = chromedp.Run(ctx,
					chromedp.Evaluate(`performance.now() - window.lastFrameTime > 16.7`, &dropRate),
				)
				if err != nil {
					t.Errorf("Failed to check frame rate: %v", err)
				}
				t.Logf("Frame drop rate during scroll: %.2f%%", dropRate*100)
			}
		})
	}
}
