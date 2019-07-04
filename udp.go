/*********************************************
                   _ooOoo_
                  o8888888o
                  88" . "88
                  (| -_- |)
                  O\  =  /O
               ____/`---'\____
             .'  \\|     |//  `.
            /  \\|||  :  |||//  \
           /  _||||| -:- |||||-  \
           |   | \\\  -  /// |   |
           | \_|  ''\---/''  |   |
           \  .-\__  `-`  ___/-. /
         ___`. .'  /--.--\  `. . __
      ."" '<  `.___\_<|>_/___.'  >'"".
     | | :  `- \`.;`\ _ /`;.`/ - ` : | |
     \  \ `-.   \_ __\ /__ _/   .-` /  /
======`-.____`-.___\_____/___.-`____.-'======
                   `=---='

^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
           佛祖保佑       永无BUG
           心外无法       法外无心
           三宝弟子       飞猪宏愿
*********************************************/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

// startUDPListener - 启动所有的UDP监听
func startUDPListener() (err error) {
	for _, name := range macrosManager.Names() {
		m := macrosManager.Get(name)
		if m.HasListenUDP() {
			err = m.StartListenUDP()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func stopUDPListener() {
	for _, name := range macrosManager.Names() {
		m := macrosManager.Get(name)
		if m.HasListenUDP() {
			m.StopListenUDP()
		}
	}
}

// HasListenUDP - 是否有监听UDP
func (m *Macro) HasListenUDP() bool {
	return m.UDP != nil && m.UDP.Port != 0
}

// StopListenUDP 停止监听UDP
func (m *Macro) StopListenUDP() {
	if !m.HasListenUDP() || m.udpconn == nil {
		return
	}

	m.udpconn.Close()

}

// StartListenUDP - 监听UDP
func (m *Macro) StartListenUDP() error {

	if !m.HasListenUDP() {
		return fmt.Errorf("%s not enable UDP Listener", m.name)
	}

	if m.udpconn != nil {
		return fmt.Errorf("%s UDP listener already start", m.name)
	}

	addr := m.UDP.IP
	if addr == "" {
		addr = "0.0.0.0"
	}

	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.ParseIP(addr),
		Port: m.UDP.Port,
	})

	if err != nil {
		return err
	}

	m.udpconn = socket

	for {
		data := make([]byte, 8192)
		dataSiz, rAddr, err := socket.ReadFromUDP(data)
		if err != nil {
			if *flagDebug > 2 {
				log.Printf(
					"%s udp listen got data error: %v",
					m.name,
					err,
				)
			}
			continue
		}

		if *flagDebug > 2 {
			log.Printf(
				"%s udp listen got %dB data: [%v] %v",
				m.name,
				dataSiz,
				data[:dataSiz],
			)
		}

		go (func(m *Macro, conn *net.UDPConn, rAddr *net.UDPAddr, data string) {

			msgdata := map[string]interface{}{}

			msgdata["data"] = string(data)
			msgHeaders := map[string]interface{}{}
			msgHeaders["remote_addr"] = rAddr.String()
			msgHeaders["remote_ip"] = rAddr.IP.String()
			msgHeaders["remote_port"] = rAddr.Port

			msgdata["__header__"] = msgHeaders

			hasReply, outMsg, _, err := m.MsgCall(msgdata)
			if err != nil {
				log.Printf("%s handle udp error: %+v\n===message===\n%v\n", m.name, err, msgdata["data"])
				return
			}

			if hasReply {

				if err != nil {
					log.Printf("%s consume message reply error: %v", m.name, err)
					return
				}

				jsonData, _ := json.Marshal(outMsg)
				_, err = conn.WriteToUDP([]byte(jsonData), rAddr)

				if err != nil {
					log.Printf("%s send udp to %s error: %v", m.name, rAddr.String(), err)
				}
			}

		})(m, socket, rAddr, string(data[:dataSiz]))

	}

}
