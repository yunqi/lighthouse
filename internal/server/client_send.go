package server

func (c *client) writeConn() {
	for p := range c.out {
		err := c.packetWriter.WritePacketAndFlush(p)
		if err != nil {
			return
		}
	}
	c.log.Debug("写入操作退出")
}
