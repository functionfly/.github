from playwright.sync_api import sync_playwright

with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    page = browser.new_page()

    page.goto('http://localhost:3000/playground')
    page.wait_for_load_state('networkidle')
    page.wait_for_timeout(3000)

    # Take screenshot
    page.screenshot(path='/tmp/playground-dark.png', full_page=True)

    # Get the page URL
    print(f"Page URL: {page.url}")

    # Check the monaco editor background (we know this works)
    monaco_bg = page.evaluate("""
        (() => {
            const monaco = document.querySelector('.monaco-editor');
            if (!monaco) return 'not found';
            const style = getComputedStyle(monaco);
            return style.backgroundColor;
        })()
    """)
    print(f"Monaco editor background: {monaco_bg}")

    # Check the Navbar background (should be dark)
    navbar = page.evaluate("""
        (() => {
            const nav = document.querySelector('nav');
            if (!nav) return 'not found';
            const style = getComputedStyle(nav);
            return {
                bg: style.backgroundColor,
                borderBottom: style.borderBottomColor
            };
        })()
    """)
    print(f"Navbar: {navbar}")

    # Check an h1 text color
    h1 = page.evaluate("""
        (() => {
            const el = document.querySelector('h1');
            if (!el) return 'not found';
            const style = getComputedStyle(el);
            return style.color;
        })()
    """)
    print(f"H1 text color: {h1}")

    # Check any element with bg-bg-primary class
    bg_primary = page.evaluate("""
        (() => {
            const el = document.querySelector('[class*="bg-bg-primary"]');
            if (!el) return 'not found';
            const style = getComputedStyle(el);
            return {
                bg: style.backgroundColor,
                className: el.className.toString().substring(0, 80)
            };
        })()
    """)
    print(f"Element with bg-bg-primary: {bg_primary}")

    # Check any element with text-text-primary class
    text_primary = page.evaluate("""
        (() => {
            const el = document.querySelector('[class*="text-text-primary"]');
            if (!el) return 'not found';
            const style = getComputedStyle(el);
            return {
                color: style.color,
                className: el.className.toString().substring(0, 80)
            };
        })()
    """)
    print(f"Element with text-text-primary: {text_primary}")

    browser.close()
    print("\nScreenshot saved to /tmp/playground-dark.png")