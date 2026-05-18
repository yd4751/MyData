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
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	MaxOpenConns int
	MaxIdleConns int
}

type Database struct {
	db          *gorm.DB
	config      DBConfig
	batchQ      []interface{}
	batchMu     sync.Mutex
	flushTicker *time.Ticker
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
	ID        int64  `gorm:"primaryKey"`
	Account   string `gorm:"size:64;unique"`
	Password  string `gorm:"size:256"`
	Salt      string `gorm:"size:64"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (a *AccountData) CheckPassword(password string) bool {
	return a.Password == password+a.Salt
}

type PlayerData struct {
	ID         int64  `gorm:"primaryKey"`
	Name       string `gorm:"size:64"`
	AccountID  int64
	Level      int32
	Exp        int64
	PosX       float64
	PosY       float64
	PosZ       float64
	Health     int32
	MaxHealth  int32
	Mana       int32
	MaxMana    int32
	Stamina    int32
	MaxStamina int32
	TeamID     int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type InventoryItem struct {
	ID        int64 `gorm:"primaryKey"`
	PlayerID  int64
	ItemID    int64
	Slot      int32
	Count     int32
	Level     int32
	UpdatedAt time.Time
}

type ItemConfig struct {
	ID          int64 `gorm:"primaryKey"`
	Name        string
	Type        int32
	EffectType  int32
	EffectValue int32
	Cooldown    int32
	MaxStack    int32
	Icon        string
	Description string
	CreatedAt   time.Time
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
	ID            int64 `gorm:"primaryKey"`
	PlayerID      int64
	AchievementID int64
	Progress      int32
	Completed     bool
	UpdatedAt     time.Time
}

type GuildMember struct {
	ID       int64 `gorm:"primaryKey"`
	GuildID  int64
	PlayerID int64
	Role     int32
	JoinedAt time.Time
}

type ChatMessage struct {
	ID        int64 `gorm:"primaryKey"`
	ChannelID int64
	SenderID  int64
	Content   string `gorm:"size:512"`
	SentAt    time.Time
}

type BattleLog struct {
	ID         int64 `gorm:"primaryKey"`
	BattleID   int64
	AttackerID int64
	TargetID   int64
	SkillID    int32
	Damage     int32
	Timestamp  time.Time
}

type LoginLog struct {
	ID        int64 `gorm:"primaryKey"`
	AccountID int64
	IP        string `gorm:"size:64"`
	LoginAt   time.Time
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
