from playwright.sync_api import sync_playwright
import time
import json

def main():
    with sync_playwright() as p:
        browser = p.chromium.launch(
            headless=True, 
            args=[
                '--no-sandbox', 
                '--disable-dev-shm-usage',
                '--disable-gpu',
                '--disable-software-rasterizer',
                '--window-size=1920,1080'
            ]
        )
        
        context = browser.new_context(
            user_agent='Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
            viewport={'width': 1920, 'height': 1080},
            accept_downloads=False,
            java_script_enabled=True
        )
        
        page = context.new_page()
        
        try:
            print("正在访问淘宝搜索页面...")
            
            search_url = "https://s.taobao.com/search?q=switch"
            page.goto(search_url, timeout=60000)
            
            print("等待页面加载...")
            page.wait_for_load_state('domcontentloaded', timeout=30000)
            time.sleep(5)
            
            page.wait_for_load_state('networkidle', timeout=30000)
            time.sleep(2)
            
            print(f"页面标题: {page.title()}")
            print(f"当前URL: {page.url}")
            
            html = page.content()
            
            with open('page_content.html', 'w', encoding='utf-8') as f:
                f.write(html)
            print("页面内容已保存到 page_content.html")
            
            print("\n=== 分析页面结构 ===")
            
            items = []
            
            all_divs = page.locator('div')
            print(f"页面中共有 {all_divs.count()} 个 div 元素")
            
            product_cards = page.locator('div[data-index]')
            print(f"找到 {product_cards.count()} 个带 data-index 的元素")
            
            for i in range(min(product_cards.count(), 3)):
                card = product_cards.nth(i)
                print(f"\n卡片 {i+1} 的HTML:")
                print(card.inner_html()[:500])
            
        except Exception as e:
            print(f"发生错误: {e}")
            
        finally:
            print("\n关闭浏览器...")
            browser.close()

if __name__ == "__main__":
    main()