from playwright.sync_api import sync_playwright
import time
import os

def run():
    with sync_playwright() as p:
        # Desktop Test
        print("--- Desktop Test ---")
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        page.goto('http://localhost:5173')
        page.wait_for_load_state('networkidle')
        
        # Find the video file and click it
        video_item = page.locator('text=test_mp4.mp4')
        video_item.click()
        
        # Wait for video to load and player to initialize
        page.wait_for_selector('video')
        time.sleep(2) # Give it a moment to settle
        
        # Get initial time
        initial_time = page.evaluate("() => document.querySelector('video').currentTime")
        print(f"Initial time: {initial_time}")
        
        # Perform drag
        # Pointer swipe is on previewStage
        stage = page.locator('div[ref="previewStage"]') # Wait, ref is not a standard attribute in DOM
        # Let's find it by class or context
        stage = page.locator('.flex-1.flex.items-center.justify-center.p-2.sm\:p-4.overflow-hidden.relative')
        
        box = stage.bounding_box()
        start_x = box['x'] + box['width'] / 2
        start_y = box['y'] + box['height'] / 2
        
        print(f"Dragging from {start_x}, {start_y}")
        page.mouse.move(start_x, start_y)
        page.mouse.down()
        page.mouse.move(start_x + 100, start_y, steps=10)
        
        # Check if overlay is visible
        overlay = page.locator('text=Seeking')
        if overlay.is_visible():
            print("Seeking overlay is visible on desktop")
            delta_text = page.locator('.text-4xl.font-black.text-white').inner_text()
            print(f"Seek delta text: {delta_text}")
        else:
            print("Seeking overlay NOT visible on desktop")
            
        page.mouse.up()
        
        time.sleep(1)
        new_time = page.evaluate("() => document.querySelector('video').currentTime")
        print(f"New time after drag: {new_time}")
        
        browser.close()

        # Mobile Test
        print("\n--- Mobile Test ---")
        iphone_13 = p.devices['iPhone 13']
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(**iphone_13)
        page = context.new_page()
        page.goto('http://localhost:5173')
        page.wait_for_load_state('networkidle')
        
        # Mobile might need to navigate differently if it's a grid
        video_item = page.locator('text=test_mp4.mp4')
        video_item.click()
        
        page.wait_for_selector('video')
        time.sleep(2)
        
        initial_time = page.evaluate("() => document.querySelector('video').currentTime")
        print(f"Initial mobile time: {initial_time}")
        
        # On mobile, we use touch events
        box = page.locator('video').bounding_box() # Drag on video directly too
        start_x = box['x'] + box['width'] / 2
        start_y = box['y'] + box['height'] / 2
        
        print(f"Swiping from {start_x}, {start_y}")
        page.touchscreen.tap(start_x, start_y) # Focus?
        
        # Simulate swipe
        page.mouse.move(start_x, start_y)
        page.mouse.down()
        page.mouse.move(start_x + 100, start_y, steps=10)
        
        overlay = page.locator('text=Seeking')
        if overlay.is_visible():
            print("Seeking overlay is visible on mobile")
        else:
            print("Seeking overlay NOT visible on mobile")
            
        page.mouse.up()
        
        time.sleep(1)
        new_time = page.evaluate("() => document.querySelector('video').currentTime")
        print(f"New mobile time after swipe: {new_time}")
        
        browser.close()

if __name__ == "__main__":
    run()
