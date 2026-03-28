from playwright.sync_api import sync_playwright

with sync_playwright() as p:
    browser = p.chromium.launch()
    page = browser.new_page()
    page.goto('http://localhost:5173')
    page.wait_for_selector('header')

    # Check group role and aria-label
    group = page.locator('div[role="group"][aria-label="Layout mode"]')
    assert group.is_visible()

    # Check aria-pressed on buttons
    grid_btn = page.locator('button[aria-label="Grid View"]')
    assert grid_btn.get_attribute('aria-pressed') == 'true'

    theme_btn = page.locator('button[aria-label="Toggle dark mode"]')
    assert theme_btn.get_attribute('aria-pressed') in ['true', 'false']

    page.screenshot(path='screenshot.png')
    browser.close()
