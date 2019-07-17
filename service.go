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

	"github.com/kardianos/service"
)

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	normalMain()
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

//service_main - 启动服务
func serviceMain() {

	svcConfig := &service.Config{
		Name:        macrosManager.ServiceName(),
		DisplayName: macrosManager.ServiceName(),
		Description: macrosManager.ServiceDesc(),
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Printf("new service error: %v", err)
		return
	}

	err = s.Run()
	if err != nil {
		log.Printf("run service error: %v", err)
	}

}
