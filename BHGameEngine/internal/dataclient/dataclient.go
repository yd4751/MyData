package dataclient

import (
	"encoding/json"
	"time"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/connector"
	"github.com/openworld-server/internal/msg"
)

type DataClient struct {
	connector *connector.Connector
}

func NewDataClient(cluster *cluster.Cluster) *DataClient {
	return &DataClient{
		connector: connector.NewConnector(cluster),
	}
}

func (c *DataClient) requestToDataService(msgID uint32, data interface{}) ([]byte, error) {
	var reqData []byte
	var err error
	if data != nil {
		reqData, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	respMsgID, _, respData, err := c.connector.RequestToNodeType(msg.NodeTypeData, msgID, reqData, 5*time.Second)
	if err != nil {
		return nil, err
	}

	if respMsgID != msgID {
		return nil, nil
	}

	return respData, nil
}

func (c *DataClient) GetAccount(account string) (*msg.AccountData, error) {
	req := map[string]string{"account": account}
	data, err := c.requestToDataService(msg.MSG_DB_ACCOUNT_GET, req)
	if err != nil {
		return nil, err
	}

	var resp msg.DBResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Result != 0 {
		return nil, nil
	}

	accountData := &msg.AccountData{}
	jsonData, _ := json.Marshal(resp.Data)
	json.Unmarshal(jsonData, accountData)
	return accountData, nil
}

func (c *DataClient) CreateAccount(account, password string) error {
	req := map[string]string{"account": account, "password": password}
	data, err := c.requestToDataService(msg.MSG_DB_ACCOUNT_CREATE, req)
	if err != nil {
		return err
	}

	var resp msg.DBResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return err
	}

	if resp.Result != 0 {
		return nil
	}
	return nil
}

func (c *DataClient) AccountExists(account string) (bool, error) {
	req := map[string]string{"account": account}
	data, err := c.requestToDataService(msg.MSG_DB_ACCOUNT_EXISTS, req)
	if err != nil {
		return false, err
	}

	var resp msg.DBResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return false, err
	}

	return resp.Result == 1, nil
}

func (c *DataClient) GetPlayerByID(playerID int64) (*msg.PlayerData, error) {
	req := map[string]int64{"player_id": playerID}
	data, err := c.requestToDataService(msg.MSG_DB_PLAYER_GET, req)
	if err != nil {
		return nil, err
	}

	var resp msg.DBResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Result != 0 {
		return nil, nil
	}

	playerData := &msg.PlayerData{}
	jsonData, _ := json.Marshal(resp.Data)
	json.Unmarshal(jsonData, playerData)
	return playerData, nil
}

func (c *DataClient) GetPlayerByAccountID(accountID int64) (*msg.PlayerData, error) {
	req := map[string]int64{"account_id": accountID}
	data, err := c.requestToDataService(msg.MSG_DB_PLAYER_GET, req)
	if err != nil {
		return nil, err
	}

	var resp msg.DBResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Result != 0 {
		return nil, nil
	}

	playerData := &msg.PlayerData{}
	jsonData, _ := json.Marshal(resp.Data)
	json.Unmarshal(jsonData, playerData)
	return playerData, nil
}

func (c *DataClient) CreatePlayer(playerID, accountID int64, name string) error {
	req := map[string]interface{}{
		"player_id":  playerID,
		"account_id": accountID,
		"name":       name,
	}
	data, err := c.requestToDataService(msg.MSG_DB_PLAYER_CREATE, req)
	if err != nil {
		return err
	}

	var resp msg.DBResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return err
	}

	if resp.Result != 0 {
		return nil
	}
	return nil
}

func (c *DataClient) UpdatePlayer(playerData *msg.PlayerData) error {
	data, err := c.requestToDataService(msg.MSG_DB_PLAYER_UPDATE, playerData)
	if err != nil {
		return err
	}

	var resp msg.DBResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return err
	}

	if resp.Result != 0 {
		return nil
	}
	return nil
}

func (c *DataClient) UpdatePlayerPosition(playerID int64, posX, posY float64) error {
	req := map[string]interface{}{
		"player_id": playerID,
		"pos_x":     posX,
		"pos_y":     posY,
	}
	data, err := c.requestToDataService(msg.MSG_DB_PLAYER_UPDATE, req)
	if err != nil {
		return err
	}

	var resp msg.DBResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return err
	}

	if resp.Result != 0 {
		return nil
	}
	return nil
}

func (c *DataClient) GetPlayerInventory(playerID int64) ([]*msg.InventoryItem, error) {
	req := map[string]int64{"player_id": playerID}
	data, err := c.requestToDataService(msg.MSG_DB_INVENTORY_GET, req)
	if err != nil {
		return nil, err
	}

	var resp msg.DBResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Result != 0 {
		return nil, nil
	}

	var items []*msg.InventoryItem
	jsonData, _ := json.Marshal(resp.Data)
	json.Unmarshal(jsonData, &items)
	return items, nil
}

func (c *DataClient) GetItemConfig(itemID int64) (*msg.ItemConfig, error) {
	req := map[string]int64{"item_id": itemID}
	data, err := c.requestToDataService(msg.MSG_DB_ITEM_GET, req)
	if err != nil {
		return nil, err
	}

	var resp msg.DBResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Result != 0 {
		return nil, nil
	}

	itemConfig := &msg.ItemConfig{}
	jsonData, _ := json.Marshal(resp.Data)
	json.Unmarshal(jsonData, itemConfig)
	return itemConfig, nil
}
