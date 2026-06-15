"""Backlink Checker - Check if URLs link to a domain."""
import re
from urllib.request import urlopen
from urllib.error import URLError, HTTPError
from html.parser import HTMLParser
import ssl


class LinkExtractor(HTMLParser):
    def __init__(self):
        super().__init__()
        self.links = []
        self.anchor_texts = []

    def handle_starttag(self, tag, attrs):
        if tag == 'a':
            href = dict(attrs).get('href', '')
            if href:
                self.links.append(href)

    def handle_data(self, data):
        data = data.strip()
        if data:
            self.anchor_texts.append(data)


def extract_links_and_anchors(html_content):
    parser = LinkExtractor()
    try:
        parser.feed(html_content)
    except Exception:
        pass
    return parser.links, parser.anchor_texts


def check_domain_in_links(links, domain_to_check):
    domain_lower = domain_to_check.lower()
    matching_links = []
    anchor_texts = []

    for link in links:
        link_lower = link.lower()
        if domain_lower in link_lower:
            matching_links.append(link)

    return matching_links


def handler(event):
    try:
        page_url = event.get("page_url", "")
        domain_to_check = event.get("domain_to_check", "")

        if not page_url:
            return {"ok": False, "error": "page_url is required"}
        if not domain_to_check:
            return {"ok": False, "error": "domain_to_check is required"}

        context = ssl.create_default_context()
        context.check_hostname = False
        context.verify_mode = ssl.CERT_NONE

        try:
            with urlopen(page_url, timeout=10, context=context) as response:
                html_content = response.read().decode('utf-8', errors='ignore')
        except HTTPError as e:
            return {"ok": False, "error": f"HTTP error: {e.code}"}
        except URLError as e:
            return {"ok": False, "error": f"URL error: {e.reason}"}
        except Exception as e:
            return {"ok": False, "error": f"Failed to fetch page: {str(e)}"}

        links, all_anchor_texts = extract_links_and_anchors(html_content)

        linking_urls = check_domain_in_links(links, domain_to_check)
        link_count = len(linking_urls)

        anchor_texts = []
        for link in linking_urls:
            for i, full_link in enumerate(links):
                if full_link == link and i < len(all_anchor_texts):
                    anchor_texts.append(all_anchor_texts[i])
                    break
            else:
                anchor_texts.append("")

        has_backlink = link_count > 0

        return {
            "ok": True,
            "has_backlink": has_backlink,
            "link_count": link_count,
            "linking_urls": linking_urls,
            "anchor_texts": anchor_texts,
            "page_url": page_url,
            "domain_checked": domain_to_check
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
