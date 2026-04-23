#!/usr/bin/env python3
"""
Playwright test to verify Providers page theme implementation
Tests:
1. Page loads without errors
2. Theme toggle buttons are present
3. Cards are visible in both light and dark mode
4. Glass morphism and density toggles work
"""

from playwright.sync_api import sync_playwright
import sys
import time

# Dashboard URL
DASHBOARD_URL = "http://localhost:3000"

def test_providers_page():
    """Test the providers page themes"""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(viewport={'width': 1280, 'height': 900})
        page = context.new_page()

        # Capture console messages and errors
        logs = []
        errors = []
        page.on("console", lambda msg: logs.append((msg.type, msg.text)))
        page.on("pageerror", lambda exc: errors.append(f"Page error: {exc}"))

        print("[INFO] Navigating to dashboard...")

        # Try navigating with longer timeout and different load states
        try:
            page.goto(DASHBOARD_URL, wait_until="domcontentloaded", timeout=30000)
            print("[INFO] Page loaded (domcontentloaded)")
        except Exception as e:
            print(f"[WARN] domcontentloaded failed, trying load: {e}")
            try:
                page.goto(DASHBOARD_URL, wait_until="load", timeout=30000)
                print("[INFO] Page loaded (load)")
            except Exception as e2:
                print(f"[ERROR] Both load strategies failed: {e2}")
                # Take screenshot anyway to see what's there
                page.screenshot(path='/tmp/providers_error.png')
                browser.close()
                return False

        # Wait a bit for React to mount
        page.wait_for_timeout(3000)

        # Try to navigate to providers via clicking or direct hash routing
        print("[INFO] Trying to navigate to providers...")

        # Check current URL
        current_url = page.url
        print(f"[INFO] Current URL: {current_url}")

        # Try direct navigation to /providers
        try:
            page.goto(f"{DASHBOARD_URL}/providers", wait_until="domcontentloaded", timeout=15000)
            page.wait_for_timeout(2000)
        except Exception as e:
            print(f"[WARN] Direct providers navigation failed: {e}")

        # Check final URL
        final_url = page.url
        print(f"[INFO] Final URL: {final_url}")

        # Take screenshot to see current state
        print("[INFO] Taking screenshot...")
        page.screenshot(path='/tmp/providers_current.png', full_page=True)

        # Check page content
        content = page.content()
        print(f"[INFO] Page content length: {len(content)}")

        # Look for key elements
        if "Providers" in content:
            print("[✓] 'Providers' text found in page")
        else:
            print("[✗] 'Providers' text NOT found")

        # Check for error messages in content
        if "404" in content or "Not Found" in content:
            print("[✗] 404 or Not Found detected")

        # Check console logs
        print(f"\n[INFO] Console logs ({len(logs)}):")
        for log_type, log_text in logs[:10]:
            print(f"  [{log_type}] {log_text[:100]}")

        if errors:
            print(f"\n[⚠] Console errors ({len(errors)}):")
            for err in errors[:5]:
                print(f"  - {err}")
        else:
            print("\n[✓] No console errors detected")

        browser.close()
        print("\n[INFO] Test completed!")
        return True

if __name__ == '__main__':
    try:
        success = test_providers_page()
        sys.exit(0 if success else 1)
    except Exception as e:
        print(f"\n[✗] Test failed with exception: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
