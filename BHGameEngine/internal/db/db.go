package db

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBConfig struct {
	Host         string // 数据库主机地址
	Port         int    // 数据库端口
	User         string // 数据库用户名
	Password     string // 数据库密码
	DBName       string // 数据库名称
	MaxOpenConns int    // 最大打开连接数
	MaxIdleConns int    // 最大空闲连接数
}

type Database struct {
	db          *gorm.DB      // GORM数据库实例
	config      DBConfig      // 数据库配置
	batchQ      []interface{} // 批量操作队列
	batchMu     sync.Mutex    // 批量操作互斥锁
	flushTicker *time.Ticker  // 批量刷新定时器
}

func NewDatabase(config DBConfig) (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Minute * 3)

	database := &Database{
		db:     db,
		config: config,
		batchQ: make([]interface{}, 0),
	}

	database.startFlushTicker()

	return database, nil
}

func (d *Database) startFlushTicker() {
	d.flushTicker = time.NewTicker(5 * time.Second)
	go func() {
		for range d.flushTicker.C {
			d.FlushBatch()
		}
	}()
}

func (d *Database) Close() error {
	d.flushTicker.Stop()
	d.FlushBatch()
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *Database) GetDB() *gorm.DB {
	return d.db
}

func (d *Database) Insert(obj interface{}) error {
	return d.db.Create(obj).Error
}

func (d *Database) Update(obj interface{}) error {
	return d.db.Save(obj).Error
}

func (d *Database) Delete(obj interface{}) error {
	return d.db.Delete(obj).Error
}

func (d *Database) First(out interface{}, where ...interface{}) error {
	return d.db.First(out, where...).Error
}

func (d *Database) Find(out interface{}, where ...interface{}) error {
	return d.db.Find(out, where...).Error
}

func (d *Database) Raw(sql string, args ...interface{}) *gorm.DB {
	return d.db.Raw(sql, args...)
}

func (d *Database) Transaction(fn func(tx *gorm.DB) error) error {
	return d.db.Transaction(fn)
}

func (d *Database) AddToBatch(obj interface{}) {
	d.batchMu.Lock()
	d.batchQ = append(d.batchQ, obj)
	if len(d.batchQ) >= 1000 {
		go d.FlushBatch()
	}
	d.batchMu.Unlock()
}

func (d *Database) FlushBatch() {
	d.batchMu.Lock()
	if len(d.batchQ) == 0 {
		d.batchMu.Unlock()
		return
	}

	batch := make([]interface{}, len(d.batchQ))
	copy(batch, d.batchQ)
	d.batchQ = make([]interface{}, 0)
	d.batchMu.Unlock()

	for _, obj := range batch {
		if err := d.db.Save(obj).Error; err != nil {
			fmt.Println("Batch save error:", err)
		}
	}
}

func (d *Database) ForceFlush() {
	d.FlushBatch()
}

type AccountData struct {
	ID        int64     `gorm:"primaryKey"`     // 账号ID
	Account   string    `gorm:"size:64;unique"` // 账号名
	Password  string    `gorm:"size:256"`       // 密码（已加密）
	Salt      string    `gorm:"size:64"`        // 加密盐
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

func (a *AccountData) CheckPassword(password string) bool {
	return a.Password == password+a.Salt
}

type PlayerData struct {
	ID         int64     `gorm:"primaryKey"` // 玩家ID
	Name       string    `gorm:"size:64"`    // 玩家名称
	AccountID  int64     // 所属账号ID
	Level      int32     // 玩家等级
	Exp        int64     // 当前经验值
	PosX       float64   // X坐标
	PosY       float64   // Y坐标
	PosZ       float64   // Z坐标
	Health     int32     // 当前生命值
	MaxHealth  int32     // 最大生命值
	Mana       int32     // 当前魔法值
	MaxMana    int32     // 最大魔法值
	Stamina    int32     // 当前体力值
	MaxStamina int32     // 最大体力值
	TeamID     int64     // 队伍ID
	CreatedAt  time.Time // 创建时间
	UpdatedAt  time.Time // 更新时间
}

type InventoryItem struct {
	ID        int64     `gorm:"primaryKey"` // 物品实例ID
	PlayerID  int64     // 所属玩家ID
	ItemID    int64     // 物品配置ID
	Slot      int32     // 背包槽位
	Count     int32     // 物品数量
	Level     int32     // 物品强化等级
	UpdatedAt time.Time // 更新时间
}

type ItemConfig struct {
	ID          int64     `gorm:"primaryKey"` // 物品配置ID
	Name        string    // 物品名称
	Type        int32     // 物品类型
	EffectType  int32     // 效果类型
	EffectValue int32     // 效果值
	Cooldown    int32     // 冷却时间(秒)
	MaxStack    int32     // 最大堆叠数量
	Icon        string    // 图标路径
	Description string    // 物品描述
	CreatedAt   time.Time // 创建时间
}

type QuestProgress struct {
	ID        int64 `gorm:"primaryKey"`
	PlayerID  int64
	QuestID   int64
	Progress  int32
	Rewarded  bool
	UpdatedAt time.Time
}

type AchievementData struct {
	ID            int64     `gorm:"primaryKey"` // 成就数据ID
	PlayerID      int64     // 玩家ID
	AchievementID int64     // 成就ID
	Progress      int32     // 当前进度
	Completed     bool      // 是否已完成
	UpdatedAt     time.Time // 更新时间
}

type GuildMember struct {
	ID       int64     `gorm:"primaryKey"` // 公会成员ID
	GuildID  int64     // 公会ID
	PlayerID int64     // 玩家ID
	Role     int32     // 职位(0-普通成员, 1-官员, 2-会长)
	JoinedAt time.Time // 加入时间
}

type ChatMessage struct {
	ID        int64     `gorm:"primaryKey"` // 聊天消息ID
	ChannelID int64     // 频道ID
	SenderID  int64     // 发送者ID
	Content   string    `gorm:"size:512"` // 消息内容
	SentAt    time.Time // 发送时间
}

type BattleLog struct {
	ID         int64     `gorm:"primaryKey"` // 战斗日志ID
	BattleID   int64     // 战斗ID
	AttackerID int64     // 攻击者ID
	TargetID   int64     // 目标ID
	SkillID    int32     // 技能ID(0表示普通攻击)
	Damage     int32     // 造成伤害
	Timestamp  time.Time // 时间戳
}

type LoginLog struct {
	ID        int64     `gorm:"primaryKey"` // 登录日志ID
	AccountID int64     // 账号ID
	IP        string    `gorm:"size:64"` // 登录IP
	LoginAt   time.Time // 登录时间
}

func (d *Database) InitTables() error {
	tables := []interface{}{
		&AccountData{},
		&PlayerData{},
		&ItemConfig{},
		&InventoryItem{},
		&QuestProgress{},
		&AchievementData{},
		&GuildMember{},
		&ChatMessage{},
		&BattleLog{},
		&LoginLog{},
	}

	for _, table := range tables {
		if err := d.db.AutoMigrate(table); err != nil {
			return err
		}
	}

	return d.InitDefaultItems()
}

func (d *Database) InitDefaultItems() error {
	var count int64
	err := d.db.Model(&ItemConfig{}).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	defaultItems := []*ItemConfig{
		{ID: 1001, Name: "小型生命药水", Type: 1, EffectType: 1, EffectValue: 50, Cooldown: 0, MaxStack: 20, Icon: "potion_red_small", Description: "恢复50点生命值"},
		{ID: 1002, Name: "中型生命药水", Type: 1, EffectType: 1, EffectValue: 100, Cooldown: 0, MaxStack: 20, Icon: "potion_red_medium", Description: "恢复100点生命值"},
		{ID: 1003, Name: "大型生命药水", Type: 1, EffectType: 1, EffectValue: 200, Cooldown: 0, MaxStack: 20, Icon: "potion_red_large", Description: "恢复200点生命值"},
		{ID: 2001, Name: "小型魔法药水", Type: 1, EffectType: 2, EffectValue: 30, Cooldown: 0, MaxStack: 20, Icon: "potion_blue_small", Description: "恢复30点魔法值"},
		{ID: 2002, Name: "中型魔法药水", Type: 1, EffectType: 2, EffectValue: 60, Cooldown: 0, MaxStack: 20, Icon: "potion_blue_medium", Description: "恢复60点魔法值"},
		{ID: 2003, Name: "大型魔法药水", Type: 1, EffectType: 2, EffectValue: 120, Cooldown: 0, MaxStack: 20, Icon: "potion_blue_large", Description: "恢复120点魔法值"},
		{ID: 3001, Name: "力量药剂", Type: 1, EffectType: 3, EffectValue: 10, Cooldown: 300, MaxStack: 10, Icon: "potion_orange", Description: "力量+10，持续5分钟"},
		{ID: 3002, Name: "敏捷药剂", Type: 1, EffectType: 4, EffectValue: 10, Cooldown: 300, MaxStack: 10, Icon: "potion_yellow", Description: "敏捷+10，持续5分钟"},
		{ID: 3003, Name: "智力药剂", Type: 1, EffectType: 5, EffectValue: 10, Cooldown: 300, MaxStack: 10, Icon: "potion_purple", Description: "智力+10，持续5分钟"},
		{ID: 4001, Name: "急救绷带", Type: 1, EffectType: 1, EffectValue: 30, Cooldown: 60, MaxStack: 10, Icon: "bandage", Description: "恢复30点生命值，冷却1分钟"},
	}

	return d.db.Create(defaultItems).Error
}

func (d *Database) GetItemConfig(itemID int64) (*ItemConfig, error) {
	var config ItemConfig
	err := d.db.First(&config, itemID).Error
	return &config, err
}

func (d *Database) GetAllItemConfigs() ([]*ItemConfig, error) {
	var configs []*ItemConfig
	err := d.db.Find(&configs).Error
	return configs, err
}

func (d *Database) CreatePlayer(playerID, accountID int64, name string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		playerData := &PlayerData{
			ID:        playerID,
			Name:      name,
			AccountID: accountID,
			Level:     1,
			Exp:       0,
			Health:    100,
			MaxHealth: 100,
			Mana:      50,
			MaxMana:   50,
			PosX:      0,
			PosY:      0,
			PosZ:      0,
		}
		if err := tx.Create(playerData).Error; err != nil {
			return err
		}

		defaultInventory := []*InventoryItem{
			{PlayerID: playerID, ItemID: 1001, Slot: 0, Count: 5},
			{PlayerID: playerID, ItemID: 1002, Slot: 1, Count: 2},
			{PlayerID: playerID, ItemID: 2001, Slot: 2, Count: 5},
			{PlayerID: playerID, ItemID: 4001, Slot: 3, Count: 3},
		}

		return tx.Create(defaultInventory).Error
	})
}

func (d *Database) AccountExists(account string) (bool, error) {
	var count int64
	err := d.db.Model(&AccountData{}).Where("account = ?", account).Count(&count).Error
	return count > 0, err
}

func (d *Database) GetAccountByAccount(account string) (*AccountData, error) {
	var data AccountData
	err := d.db.Where("account = ?", account).First(&data).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &data, nil
}

func (d *Database) CreateAccount(account, password string) error {
	salt := generateSalt()
	hashedPassword := hashPassword(password, salt)
	data := &AccountData{
		Account:  account,
		Password: hashedPassword,
		Salt:     salt,
	}
	return d.db.Create(data).Error
}

func (d *Database) GetPlayerByAccountID(accountID int64) (*PlayerData, error) {
	var data PlayerData
	err := d.db.Where("account_id = ?", accountID).First(&data).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &data, nil
}

func (d *Database) GetPlayerByID(playerID int64) (*PlayerData, error) {
	var data PlayerData
	err := d.db.First(&data, playerID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &data, nil
}

func (d *Database) UpdatePlayerPosition(playerID int64, posX, posY float64) error {
	return d.db.Model(&PlayerData{}).Where("id = ?", playerID).UpdateColumns(map[string]interface{}{
		"pos_x": posX,
		"pos_y": posY,
	}).Error
}

func generateSalt() string {
	return "salt"
}

func hashPassword(password, salt string) string {
	return password + salt
}

func (d *Database) GetPlayerData(playerID int64) (*PlayerData, error) {
	var data PlayerData
	err := d.db.First(&data, playerID).Error
	return &data, err
}

func (d *Database) SavePlayerData(data *PlayerData) error {
	return d.db.Save(data).Error
}

func (d *Database) GetPlayerInventory(playerID int64) ([]*InventoryItem, error) {
	var items []*InventoryItem
	err := d.db.Where("player_id = ?", playerID).Find(&items).Error
	return items, err
}

func (d *Database) GetPlayerIDByName(name string) (int64, error) {
	var data PlayerData
	err := d.db.Where("name = ?", name).First(&data).Error
	if err != nil {
		return 0, err
	}
	return data.ID, nil
}

func (d *Database) AddItemToInventory(playerID int64, itemID int64, count int32) error {
	var existingItem InventoryItem
	err := d.db.Where("player_id = ? AND item_id = ?", playerID, itemID).First(&existingItem).Error

	if err == gorm.ErrRecordNotFound {
		var maxSlot *int32
		err := d.db.Model(&InventoryItem{}).Where("player_id = ?", playerID).Select("MAX(slot)").Scan(&maxSlot).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		slot := int32(1)
		if maxSlot != nil {
			slot = *maxSlot + 1
		}

		newItem := &InventoryItem{
			PlayerID: playerID,
			ItemID:   itemID,
			Slot:     slot,
			Count:    count,
		}
		return d.db.Create(newItem).Error
	} else if err != nil {
		return err
	}

	return d.db.Model(&existingItem).Update("count", existingItem.Count+count).Error
}

func (d *Database) GetQuestProgress(playerID int64) ([]*QuestProgress, error) {
	var progresses []*QuestProgress
	err := d.db.Where("player_id = ?", playerID).Find(&progresses).Error
	return progresses, err
}

func (d *Database) RecordBattleLog(log *BattleLog) error {
	return d.db.Create(log).Error
}

func (d *Database) RecordLogin(accountID int64, ip string) error {
	log := &LoginLog{
		AccountID: accountID,
		IP:        ip,
		LoginAt:   time.Now(),
	}
	return d.db.Create(log).Error
}
