package main

import (
	"fmt"

	"github.com/openworld-server/internal/db"
	"github.com/openworld-server/pkg/config"
)

func main() {
	err := config.LoadConfig("./config/config.toml")
	if err != nil {
		fmt.Println("Failed to load config:", err)
		return
	}

	mysqlConfig := config.GetMySQLConfig()
	database, err := db.NewDatabase(db.DBConfig{
		Host:         mysqlConfig.Host,
		Port:         mysqlConfig.Port,
		User:         mysqlConfig.User,
		Password:     mysqlConfig.Password,
		DBName:       mysqlConfig.DBName,
		MaxOpenConns: mysqlConfig.MaxOpenConns,
		MaxIdleConns: mysqlConfig.MaxIdleConns,
	})
	if err != nil {
		fmt.Println("Failed to connect to MySQL:", err)
		return
	}
	defer database.Close()

	playerName := "bighat"
	playerID, err := database.GetPlayerIDByName(playerName)
	if err != nil {
		fmt.Printf("Failed to find player %s: %v\n", playerName, err)
		return
	}
	fmt.Printf("Found player %s with ID: %d\n", playerName, playerID)

	itemsToAdd := []struct {
		ItemID int64
		Count  int32
	}{
		{1001, 10},
		{1002, 5},
		{1003, 3},
		{2001, 10},
		{2002, 5},
		{2003, 3},
		{3001, 2},
		{3002, 2},
		{3003, 2},
		{4001, 5},
	}

	for _, item := range itemsToAdd {
		err := database.AddItemToInventory(playerID, item.ItemID, item.Count)
		if err != nil {
			fmt.Printf("Failed to add item %d x %d: %v\n", item.ItemID, item.Count, err)
		} else {
			fmt.Printf("Added item %d x %d successfully\n", item.ItemID, item.Count)
		}
	}

	fmt.Println("All items added to player", playerName)
}
