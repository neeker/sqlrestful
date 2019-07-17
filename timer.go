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
	"log"
	"time"
)

// startTimers - 启动所有定时器
func startTimers() error {
	for _, name := range macrosManager.Names() {
		m := macrosManager.Get(name)
		if m.HasTimer() {
			err := m.StartTimer()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// HasTimer - 是否有定时器
func (m *Macro) HasTimer() bool {
	return m.Timer != nil && m.Timer.Inteval > 0
}

// StartTimer - 启动定时器
func (m *Macro) StartTimer() error {
	if !m.HasTimer() {
		return nil
	}

	if *flagDebug > 2 {
		log.Printf("call %s timer", m.name)
	}

	go func(m *Macro) {
		t := time.NewTicker(time.Duration(m.Timer.Inteval) * time.Second)
		for {
			if *flagDebug > 2 {
				log.Printf("wait %s timer run...", m.name)
			}
			select {
			case <-t.C:
				err := m.TimerCall()
				if err != nil {
					log.Printf("%s timer call error: %v", m.name, err)
				}
			}
		}
	}(m)
	return nil
}
