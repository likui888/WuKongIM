package logic

import (
	"sync"

	"github.com/WuKongIM/WuKongIM/internal/gatewaycommon"
	wkproto "github.com/WuKongIM/WuKongIMGoProto"
)

type clientConnManager struct {
	userConnMap map[string][]int64
	connMap     map[int64]gatewaycommon.Conn

	sync.RWMutex
}

func newClientConnManager() *clientConnManager {

	return &clientConnManager{
		userConnMap: make(map[string][]int64),
		connMap:     make(map[int64]gatewaycommon.Conn),
	}
}

func (c *clientConnManager) AddConn(conn gatewaycommon.Conn) {
	c.Lock()
	defer c.Unlock()
	connIDs := c.userConnMap[conn.UID()]
	if connIDs == nil {
		connIDs = make([]int64, 0, 10)
	}
	connIDs = append(connIDs, conn.ID())
	c.userConnMap[conn.UID()] = connIDs
	c.connMap[conn.ID()] = conn
}

func (c *clientConnManager) GetConn(id int64) gatewaycommon.Conn {
	c.RLock()
	defer c.RUnlock()
	return c.connMap[id]
}

func (c *clientConnManager) RemoveConn(conn gatewaycommon.Conn) {
	c.RemoveConnWithID(conn.ID())
}

func (c *clientConnManager) RemoveConnWithID(id int64) {
	c.Lock()
	defer c.Unlock()
	conn := c.connMap[id]
	delete(c.connMap, id)
	if conn == nil {
		return
	}
	connIDs := c.userConnMap[conn.UID()]
	if len(connIDs) > 0 {
		for index, connID := range connIDs {
			if connID == conn.ID() {
				connIDs = append(connIDs[:index], connIDs[index+1:]...)
				c.userConnMap[conn.UID()] = connIDs
			}
		}
	}
}

func (c *clientConnManager) GetConnsWithUID(uid string) []gatewaycommon.Conn {
	c.RLock()
	defer c.RUnlock()
	connIDs := c.userConnMap[uid]
	if len(connIDs) == 0 {
		return nil
	}
	conns := make([]gatewaycommon.Conn, 0, len(connIDs))
	for _, id := range connIDs {
		conn := c.connMap[id]
		if conn != nil {
			conns = append(conns, conn)
		}
	}
	return conns
}

func (c *clientConnManager) ExistConnsWithUID(uid string) bool {
	c.RLock()
	defer c.RUnlock()
	return len(c.userConnMap[uid]) > 0
}

func (c *clientConnManager) GetConnsWith(uid string, deviceFlag wkproto.DeviceFlag) []gatewaycommon.Conn {
	conns := c.GetConnsWithUID(uid)
	if len(conns) == 0 {
		return nil
	}
	deviceConns := make([]gatewaycommon.Conn, 0, len(conns))
	for _, conn := range conns {
		if conn.DeviceFlag() == deviceFlag.ToUint8() {
			deviceConns = append(deviceConns, conn)
		}
	}
	return deviceConns
}

// GetConnCountWith 获取设备的在线数量和用户所有设备的在线数量
func (c *clientConnManager) GetConnCountWith(uid string, deviceFlag wkproto.DeviceFlag) (int, int) {
	conns := c.GetConnsWithUID(uid)
	if len(conns) == 0 {
		return 0, 0
	}
	deviceOnlineCount := 0
	for _, conn := range conns {
		if wkproto.DeviceFlag(conn.DeviceFlag()) == deviceFlag {
			deviceOnlineCount++
		}
	}
	return deviceOnlineCount, len(conns)
}

// GetOnlineConns 传一批uids 返回在线的uids
func (c *clientConnManager) GetOnlineConns(uids []string) []gatewaycommon.Conn {
	if len(uids) == 0 {
		return make([]gatewaycommon.Conn, 0)
	}
	c.Lock()
	defer c.Unlock()
	var onlineConns = make([]gatewaycommon.Conn, 0, len(uids))
	for _, uid := range uids {
		connIDs := c.userConnMap[uid]
		for _, connID := range connIDs {
			conn := c.connMap[connID]
			if conn != nil {
				onlineConns = append(onlineConns, conn)
			}
		}
	}
	return onlineConns
}

func (c *clientConnManager) GetAllConns() []gatewaycommon.Conn {
	c.RLock()
	defer c.RUnlock()
	conns := make([]gatewaycommon.Conn, 0, len(c.connMap))
	for _, conn := range c.connMap {
		conns = append(conns, conn)
	}
	return conns
}

func (c *clientConnManager) Reset() {
	c.Lock()
	defer c.Unlock()
	c.userConnMap = make(map[string][]int64)
	c.connMap = make(map[int64]gatewaycommon.Conn)
}