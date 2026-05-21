package main

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type News struct {
	ID          int       `db:"id"`
	Title       string    `db:"title"`
	Summary     string    `db:"summary"`
	Source      string    `db:"source"`
	PublishTime time.Time `db:"publish_time"`
	Lat         float64   `db:"lat"`
	Lng         float64   `db:"lng"`
	GeoLevel    string    `db:"geo_level"`
	IsBreaking  int       `db:"is_breaking"`
	Priority    int       `db:"priority"`
	CreatedAt   time.Time `db:"created_at"`
}

var newsData = []News{
	// Asia
	{Title: "北京科技峰会盛大开幕", Summary: "2024北京国际科技创新峰会今日在北京国家会议中心隆重开幕", Source: "新华社", Lat: 39.9042, Lng: 116.4074, GeoLevel: "city", IsBreaking: 1, Priority: 1},
	{Title: "东京股市创新高", Summary: "日经225指数今日突破38000点大关", Source: "共同社", Lat: 35.6762, Lng: 139.6503, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "首尔发布5G发展报告", Summary: "韩国首尔市发布最新5G发展报告", Source: "韩联社", Lat: 37.5665, Lng: 126.9780, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	{Title: "新加坡推出碳中和计划", Summary: "新加坡政府宣布碳中和计划", Source: "联合早报", Lat: 1.3521, Lng: 103.8198, GeoLevel: "city", IsBreaking: 1, Priority: 1},
	{Title: "孟买IT产业持续增长", Summary: "印度孟买IT产业持续快速增长", Source: "印度时报", Lat: 19.0760, Lng: 72.8777, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "香港金融科技周开幕", Summary: "2024香港金融科技周今日开幕", Source: "香港商报", Lat: 22.3193, Lng: 114.1694, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	{Title: "上海人工智能研究院成立", Summary: "上海人工智能研究院正式成立", Source: "解放日报", Lat: 31.2304, Lng: 121.4737, GeoLevel: "city", IsBreaking: 1, Priority: 1},
	{Title: "迪拜举办全球气候论坛", Summary: "迪拜成功举办全球气候论坛", Source: "海湾新闻", Lat: 25.2048, Lng: 55.2708, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "曼谷旅游旺季来临", Summary: "泰国曼谷迎来旅游旺季", Source: "曼谷邮报", Lat: 13.7563, Lng: 100.5018, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	{Title: "雅加达智慧城市建设加速", Summary: "印尼雅加达智慧城市建设加速推进", Source: "雅加达邮报", Lat: -6.2088, Lng: 106.8456, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "深圳新能源汽车销量创新高", Summary: "深圳新能源汽车销量再创新高", Source: "深圳特区报", Lat: 22.5431, Lng: 114.0579, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "台北半导体产业论坛", Summary: "2024台北半导体产业论坛召开", Source: "联合报", Lat: 25.0330, Lng: 121.5654, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	{Title: "马尼拉数字经济峰会", Summary: "菲律宾马尼拉举办数字经济峰会", Source: "菲律宾星报", Lat: 14.5995, Lng: 120.9842, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "吉隆坡科技创新中心启用", Summary: "马来西亚吉隆坡科技创新中心正式启用", Source: "星洲日报", Lat: 3.1390, Lng: 101.6869, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	{Title: "河内高新技术园区扩建", Summary: "越南河内高新技术园区扩建工程启动", Source: "越南时报", Lat: 21.0278, Lng: 105.8342, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	// Europe
	{Title: "柏林人工智能大会", Summary: "2024柏林人工智能大会吸引全球专家学者", Source: "明镜周刊", Lat: 52.5200, Lng: 13.4050, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "伦敦金融科技展", Summary: "伦敦举办盛大金融科技展", Source: "金融时报", Lat: 51.5074, Lng: -0.1278, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "巴黎气候峰会", Summary: "巴黎召开国际气候峰会", Source: "世界报", Lat: 48.8566, Lng: 2.3522, GeoLevel: "city", IsBreaking: 1, Priority: 1},
	{Title: "柏林电子消费品展", Summary: "IFA 2024柏林电子消费品展盛大开幕", Source: "法兰克福汇报", Lat: 52.5200, Lng: 13.4050, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "马德里可再生能源论坛", Summary: "西班牙马德里举办可再生能源论坛", Source: "国家报", Lat: 40.4168, Lng: -3.7038, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	{Title: "罗马文化遗产保护会议", Summary: "罗马召开文化遗产保护国际会议", Source: "共和报", Lat: 41.9028, Lng: 12.4964, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "阿姆斯特丹智慧城市峰会", Summary: "阿姆斯特丹智慧城市峰会展示可持续发展成果", Source: "荷兰商报", Lat: 52.3676, Lng: 4.9041, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	{Title: "斯德哥尔摩科技周", Summary: "瑞典斯德哥尔摩科技周开幕", Source: "瑞典日报", Lat: 59.3293, Lng: 18.0686, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "维也纳国际原子能会议", Summary: "维也纳举办国际原子能会议", Source: "标准报", Lat: 48.2082, Lng: 16.3738, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	{Title: "布鲁塞尔欧盟数字战略发布会", Summary: "欧盟委员会发布最新数字战略", Source: "欧洲时报", Lat: 50.8503, Lng: 4.3517, GeoLevel: "city", IsBreaking: 1, Priority: 1},
	{Title: "哥本哈根绿色科技展", Summary: "哥本哈根绿色科技展展示环保技术", Source: "政治报", Lat: 55.6761, Lng: 12.5683, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "里斯本创业峰会", Summary: "葡萄牙里斯本举办创业峰会", Source: "公众报", Lat: 38.7223, Lng: -9.1393, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	{Title: "华沙科技创新论坛", Summary: "波兰华沙科技创新论坛聚焦人工智能", Source: "选举报", Lat: 52.2297, Lng: 21.0122, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "布拉格数字艺术展", Summary: "布拉格数字艺术展展示新媒体艺术", Source: "权利报", Lat: 50.0755, Lng: 14.4378, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	{Title: "奥斯陆可持续发展大会", Summary: "挪威奥斯陆举办可持续发展大会", Source: "晚邮报", Lat: 59.9139, Lng: 10.7522, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	// Americas
	{Title: "纽约纳斯达克创新峰会", Summary: "纽约纳斯达克举办创新峰会", Source: "华尔街日报", Lat: 40.7128, Lng: -74.0060, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "旧金山AI开发者大会", Summary: "2024旧金山AI开发者大会吸引全球开发者", Source: "旧金山纪事报", Lat: 37.7749, Lng: -122.4194, GeoLevel: "city", IsBreaking: 1, Priority: 1},
	{Title: "多伦多国际电影节", Summary: "多伦多国际电影节盛大开幕", Source: "多伦多星报", Lat: 43.6532, Lng: -79.3832, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	{Title: "洛杉矶科技展", Summary: "洛杉矶举办年度科技展", Source: "洛杉矶时报", Lat: 34.0522, Lng: -118.2437, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "芝加哥制造业峰会", Summary: "芝加哥制造业峰会聚焦智能制造", Source: "芝加哥论坛报", Lat: 41.8781, Lng: -87.6298, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "迈阿密区块链大会", Summary: "迈阿密举办全球区块链大会", Source: "迈阿密先驱报", Lat: 25.7617, Lng: -80.1918, GeoLevel: "city", IsBreaking: 1, Priority: 1},
	{Title: "西雅图云计算大会", Summary: "西雅图云计算大会汇聚全球云服务巨头", Source: "西雅图时报", Lat: 47.6062, Lng: -122.3321, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "波士顿生物科技峰会", Summary: "波士顿生物科技峰会探讨生命科学突破", Source: "波士顿环球报", Lat: 42.3601, Lng: -71.0589, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "华盛顿气候政策论坛", Summary: "华盛顿举办气候政策论坛", Source: "华盛顿邮报", Lat: 38.9072, Lng: -77.0369, GeoLevel: "city", IsBreaking: 1, Priority: 1},
	{Title: "温哥华绿色科技展", Summary: "温哥华绿色科技展展示环保技术", Source: "温哥华太阳报", Lat: 49.2827, Lng: -123.1207, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	{Title: "墨西哥城数字转型大会", Summary: "墨西哥城数字转型大会推动拉美数字化", Source: "改革报", Lat: 19.4326, Lng: -99.1332, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "圣保罗科技创新周", Summary: "巴西圣保罗科技创新周展示成果", Source: "圣保罗页报", Lat: -23.5505, Lng: -46.6333, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "布宜诺斯艾利斯创业大会", Summary: "阿根廷布宜诺斯艾利斯创业大会", Source: "号角报", Lat: -34.6037, Lng: -58.3816, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	{Title: "智利圣地亚哥新能源峰会", Summary: "智利圣地亚哥新能源峰会推动可再生能源", Source: "信使报", Lat: -33.4489, Lng: -70.6693, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "哥伦比亚波哥大智慧城市大会", Summary: "波哥大智慧城市大会探讨可持续发展", Source: "时代报", Lat: 4.6097, Lng: -74.0817, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	// Oceania
	{Title: "悉尼科技博览会", Summary: "悉尼科技博览会展示澳大利亚创新成果", Source: "悉尼先驱晨报", Lat: -33.8688, Lng: 151.2093, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "奥克兰可持续发展论坛", Summary: "新西兰奥克兰可持续发展论坛", Source: "新西兰先驱报", Lat: -36.8485, Lng: 174.7633, GeoLevel: "city", IsBreaking: 0, Priority: 3},
	// Africa
	{Title: "开普敦科技峰会", Summary: "南非开普敦科技峰会推动非洲科技创新", Source: "开普时报", Lat: -33.9249, Lng: 18.4241, GeoLevel: "city", IsBreaking: 0, Priority: 2},
	{Title: "拉各斯数字经济大会", Summary: "尼日利亚拉各斯数字经济大会", Source: "先锋报", Lat: 6.5244, Lng: 3.3792, GeoLevel: "city", IsBreaking: 0, Priority: 3},
}

func main() {
	db, err := sqlx.Connect("mysql", "root:12345678@tcp(localhost:3306)/geonews?charset=utf8mb4&parseTime=true")
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	tx, err := db.Beginx()
	if err != nil {
		fmt.Printf("Failed to begin transaction: %v\n", err)
		return
	}

	now := time.Now()
	for i, news := range newsData {
		news.CreatedAt = now
		news.PublishTime = now.Add(-time.Duration(i) * time.Hour)

		_, err := tx.NamedExec(`INSERT INTO news (title, summary, source, publish_time, lat, lng, geo_level, is_breaking, priority, created_at) 
			VALUES (:title, :summary, :source, :publish_time, :lat, :lng, :geo_level, :is_breaking, :priority, :created_at)`, news)
		if err != nil {
			tx.Rollback()
			fmt.Printf("Failed to insert news %d: %v\n", i+1, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		fmt.Printf("Failed to commit transaction: %v\n", err)
		return
	}

	fmt.Printf("Successfully inserted %d news records\n", len(newsData))
}
