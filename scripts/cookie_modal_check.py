from playwright.sync_api import sync_playwright

with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    context = browser.new_context()
    page = context.new_page()

    page.goto('http://localhost:3000', wait_until='domcontentloaded')
    page.wait_for_timeout(2000)

    print("URL after initial load:", page.url)

    # Fill in login form - looking for email and password fields
    email_input = page.locator('input[type="email"], input[name="email"], input[id="email"]').first
    password_input = page.locator('input[type="password"], input[name="password"], input[id="password"]').first

    if email_input.count() > 0 and password_input.count() > 0:
        print("Found login fields")
        email_input.fill('admin@functionfly.local')
        password_input.fill('admin123')
        page.wait_for_timeout(500)

        # Find and click sign in button
        signin_btn = page.locator('button:has-text("Sign in"), button:has-text("Continue")').first
        if signin_btn.count() > 0:
            signin_btn.click()
            page.wait_for_timeout(3000)
            print("After login URL:", page.url)
    else:
        print("No login fields found")
        # Maybe the page uses a different login method
        # Check what's on the page
        page.screenshot(path='/tmp/login_page.png', full_page=True)
        print("Screenshot saved to /tmp/login_page.png")

    # Now check for cookie consent
    cc_main_count = page.locator('#cc-main').count()
    print(f"#cc-main elements: {cc_main_count}")

    page.screenshot(path='/tmp/after_login.png', full_page=True)
    print("Screenshot saved to /tmp/after_login.png")

    browser.close()