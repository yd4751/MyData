# HTTP headers for requests
import json
import os
from urllib.parse import urlparse

from bs4 import BeautifulSoup
#文件保存路径
save_path = os.path.dirname(os.path.abspath(__file__))

headers = {
    'authority': 'www.txt808.cc',
    'method': 'GET',
    'path': '/static/api/js/trans/logger.js?v=d16ec0e3.js',
    'scheme': 'https',
    'accept': '*/*',
    'accept-encoding': 'gzip, deflate, br',
    'accept-language': 'zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,ko;q=0.6',
    'cache-control': 'no-cache',
    'pragma': 'no-cache',
    'referer': 'https://www.txt808.cc/',
    'sec-ch-ua': '"Chromium";v="104", " Not A;Brand";v="99", "Google Chrome";v="104"',
    'sec-ch-ua-mobile': '?0',
    'sec-ch-ua-platform': '"Windows"',
    'sec-fetch-dest': 'script',
    'sec-fetch-mode': 'no-cors',
    'sec-fetch-site': 'same-origin',
    'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36'
}
# Cookies for requests (in URL encoded format, used as HTTP parameters)
cookies = {
    '_pk_ref.5.a082': '%5B%22%22%2C%22%22%2C1779257071%2C%22https%3A%2F%2Fwww.baidu.com%2Flink%3Furl%3DSOIBDq4pJmm5GKn1LjofrTOA61xZs1wH6_5rxannRqq%26wd%3D%26eqid%3Df8de6b3d09a991ce000000066a0d4ee8%22%5D',
    '_pk_id.5.a082': '3cd00773a487e215.1779257071.',
    '_pk_ses.5.a082': '1'
}
#
def format_path(file_name):
    return os.path.join(save_path, file_name)
#
def task(engine):
    global save_path
    save_path = os.path.join(save_path, "saves")
    if not os.path.exists(save_path):
        os.makedirs(save_path)

    engine.add_task(
        engine.create_task("https://www.txt808.cc","GET")
        .with_headers(headers)
        .with_cookies(cookies)
        .with_callback(handle_entry)
        .build()
    )
    #
    engine.logger.info(f"========== Task Start ==========")
    engine.logger.info(f"save path: {save_path}")
#
def handle_entry(engine,response,custom):
    #with open(format_path("entry.txt"), "wb") as file:
    #    file.write(response['content'])
    url = response['url']
    #解析
    soup = BeautifulSoup(response['content'], 'html.parser')
    r1 = soup.select("ul.menu_mid > li > a")[1].get('href')
    #print(r1)
    next_url = "{0}{1}".format(url,r1)
    #engine.logger.info(url)

    #
    new_tasks = []
    task = engine.create_task(next_url,"GET") \
                .with_headers(headers) \
                .with_cookies(cookies) \
                .with_callback(url_all) \
                .build()
    new_tasks.append(task)
    
    return new_tasks
#
def url_all(engine,response,custom):
    #with open(format_path("/all.txt"), "wb") as file:
    #    file.write(response['content'])
    '''
    <div class="slist">
    <div class="pic">
    <a href="/yanqing/txt62596.html" target="_blank">
    <img src="https://pic.8080txt.com/d/file/pic/2026/03/03/zr1v2zoneuo.jpg" width="100" height="130" Alt="《深渊女神》全本TXT下载-作者：藤萝为枝"></a>
    </div>
    <div class="info">
    <h4><a href="/yanqing/txt62596.html" target="_blank">《深渊女神》全本TXT小说下载</a><span class="fr"><b>0人下载</b></span></h4>

    <div id="splitpage">
    <span>总数<strong>30000</strong> / 共<strong>3000</strong>页</span>
    <ul><li class=active>1</li>
    <li><a href="/all/index_2.html">2</a></li>
    <li><a href="/all/index_3.html">3</a></li>
    <li><a href="/all/index_4.html">4</a></li><li><a href="/all/index_5.html">5</a></li><li><a href="/all/index_6.html">6</a></li><li><a href="/all/index_7.html">7</a></li><li><a href="/all/index_8.html">8</a></li><li><a href="/all/index_9.html">9</a></li><li><a href="/all/index_10.html">10</a></li><li class=next><a href="/all/index_2.html">下一页</a></li><li class=lastly><a href="/all/index_3000.html">尾页</a></li></ul></ul>
    </div>
    '''
    parsed_url = urlparse(response['url'])
    domain = f"{parsed_url.scheme}://{parsed_url.netloc}"
    new_tasks = []
    #列表
    soup = BeautifulSoup(response['content'], 'html.parser')
    cur_list = soup.select("div.slist > div.info > h4 > a")
    for item in cur_list:
        url = "{0}{1}".format(domain,item.get('href'))
        engine.logger.info(url)

        task = engine.create_task(url,"GET") \
                    .with_headers(headers) \
                    .with_cookies(cookies) \
                    .with_callback(get_item_page) \
                    .build()
        new_tasks.append(task) 
        break
    #下一页
    #...
    return new_tasks
#
def get_item_page(engine,response,custom):
    #with open(format_path("{}.txt".format(os.path.basename(response['url']))), "wb") as file:
    #    file.write(response['content'])

    '''
    <div class="nrlist">
    <dl>
	<dt class="pic fl"><img class="pics3" src="https://pic.8080txt.com/d/file/pic/2026/03/03/zr1v2zoneuo.jpg" alt="深渊女神图片">
    <h2>深渊女神</h2>
    <dd class="db">小说作者：<a href="http://www.8080txt.com/writer/藤萝为枝/"  title="藤萝为枝">藤萝为枝</a></dd>
    <dd class="db">小说状态：<span>完结</span></dd>
    <dd class="db">小说分类：<a href="/yanqing/" title="女生言情">女生言情</a></dd>
    <dd class="db">授权媒体：<span>网络转载</span></dd>
    <dd class="db">小说格式：TXT电子书</dd>
    <dd class="db">小说大小：<span>1.2 MB</span></dd>
    <dd class="db">发布时间: <span>2026-05-20</span></dd>

    <div class="softsayxq">
	<div class="cont">喻嗔长得好看，还生来带香。微笑唇一弯甜蜜毓秀。不出意外，她这辈子妥妥拿的女主剧本。可是为了报答柏正，她被迫换成了倒霉苦情女配剧本。体校柏正张狂不羁，人憎狗嫌，干的都是混账事。柏少这辈子最讨厌的三种人。穷鬼、天真、长得好。喻嗔家乡倒塌一无所有，穷。他们小镇的人都单纯。她很漂亮，是柏少生平未见过的美声明：全集TXT小说<a href="/yanqing/txt62596.html" target="_blank"><font color="000">《深渊女神》</font></a>由书友上传至<a href="http://www.8080txt.com" target="_blank"><font color="000">80电子书</font></a>，该作品内容版权归原作者、出版社所有。如原作者、出版社认为本站行为侵权，请联系本站，本站会立即删除您认为侵权的作品。

    <div class="downlinks">
    <ul>
	<li>
	<p>
	<b><a href="/down/txt2c62596b0.html" target="_blank">进入小说下载地址</a></b>
    '''
    parsed_url = urlparse(response['url'])
    domain = f"{parsed_url.scheme}://{parsed_url.netloc}"
    #基本数据
    soup = BeautifulSoup(response['content'], 'html.parser')
    r1 = soup.select_one("div.nrlist > dl")
    r2 = r1.select("dd.db")
    r3 = soup.select_one("div.softsayxq > div.cont")
    r4 = soup.select_one("div.downlinks > ul > li > p > b > a")
    custom = {
        "img": r1.select_one("dt.pic > img").get('src'),
        "title": r1.select_one("h2").get_text(),
        "author": r2[0].select_one("a").get_text(),
        "status": r2[1].select_one("span").get_text(),
        "category": r2[2].select_one("a").get_text(),
        "media": r2[3].select_one("span").get_text(),
        "format": r2[4].get_text().replace("小说格式：", "").strip(),
        "size": r2[5].select_one("span").get_text(),
        "pub_date": r2[6].select_one("span").get_text(),
        "content":r3.get_text(),
        "down_url": domain + r4.get('href')
    }
    #engine.logger.info(down_url)
    #
    tasks = []
    task = engine.create_task(custom["down_url"],"GET") \
                .with_headers(headers) \
                .with_cookies(cookies) \
                .with_callback(download_page) \
                .with_custom(custom) \
                .build()
    tasks.append(task)
    return tasks
#
def download_page(engine,response,custom):
    #with open(format_path("{}.txt".format(os.path.basename(response['url']))), "wb") as file:
    #    file.write(response['content'])
    #print(custom)
    '''
    <div class="downlist">
    <li>&nbsp;下载地址1：<strong><a href="https://down.8080txt.com/d/file/down/2026/03/03/深渊女神.txt" title="《深渊女神》全集TXT下载" download="[txt808.cc]深渊女神"><font color="#ff0000">下载到电脑（右键另存）</font></a></strong></li>
    </div>
    <div class="downlist">
    <li>&nbsp;下载地址2：<strong><a href="https://down.txt8080.com/d/file/down/2026/03/03/深渊女神.txt" title="《深渊女神》全集TXT下载" download="[txt808.cc]深渊女神"><font color="#ff0000">下载到电脑（右键另存）</font></a></strong></li>
    </div>
    <div class="downlist">
    <li>&nbsp;下载地址3：<strong><a rel="nofollow" href="/wz.php" target="_blank"><font color="#ff0000">更多TXT小说下载网站</font></a></strong></li>
    </div>
    '''
    soup = BeautifulSoup(response['content'], 'html.parser')
    soup = soup.select("div.downlist > li > strong > a")
    tasks = []
    for item in soup:
        if ".txt" in item.get('href'):
            task = engine.create_task(item.get('href'),"GET") \
                        .with_headers(headers) \
                        .with_cookies(cookies) \
                        .with_callback(download_file) \
                        .with_custom(custom) \
                        .build()
            tasks.append(task)
            break
    #img
    task = engine.create_task(custom["img"],"GET") \
                        .with_headers(headers) \
                        .with_cookies(cookies) \
                        .with_callback(download_file) \
                        .with_custom(custom) \
                        .build()
    tasks.append(task)
    
    return tasks

#
def download_file(engine,response,custom):
    engine.logger.info(f"Save file {response['status_code']}, length: {len(response['content'])} name: {custom['title']} url: {response['url']}")    
    ext = response['url'].split('.')[-1]
    #内容
    with open(format_path("{0}.{1}".format(custom['title'],ext)), "wb") as file:
        file.write(response['content'])

    return None