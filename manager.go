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
	"fmt"
	"strings"

	"io/ioutil"
	"log"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/dgrijalva/jwt-go"

	"github.com/hashicorp/hcl"
)

// Manager - a macros manager
type Manager struct {
	meta     *Macro
	macros   map[string]*Macro
	compiled *template.Template
	sync.RWMutex
}

func fixMacro(v *Macro) {
	if len(v.Total) > 0 {
		v.Result = "page"
	}

	if v.Result == "" {
		v.Result = "list"
	} else {
		v.Result = strings.ToLower(v.Result)
	}

	if v.Impl == "" {
		v.Impl = "sql"
	} else {
		v.Impl = strings.ToLower(v.Impl)
	}

	if v.Impl != "js" && v.Impl != "cmd" {
		v.Impl = "sql"
	}

	if v.Format == "" {
		v.Format = "enclosed"
	} else {
		v.Format = strings.ToLower(v.Format)
	}

	if v.Format != "origin" && v.Format != "nil" && v.Format != "redirect" {
		v.Format = "enclosed"
	}

	switch {
	case v.Result == "list":
	case v.Result == "object":
	case v.Result == "page":
	default:
		v.Result = "list"
	}

	if v.Summary == "" {
		v.Summary = v.name
	}

	if len(v.Tags) == 0 {
		v.Tags = append(v.Tags, v.name)
	}

	if v.Security != nil {
		if v.Security.Policy == "" {
			v.Security.Policy = "allow"
		} else {
			v.Security.Policy = strings.ToLower(v.Security.Policy)
		}
		if v.Security.Policy != "allow" {
			v.Security.Policy = "deny"
		}

		if v.Security.Users != nil && len(v.Security.Users) > 0 {
			v.usersMap = map[string]bool{}
			for _, u := range v.Security.Users {
				v.usersMap[u] = true
			}
		}

		if v.Security.Roles != nil && len(v.Security.Roles) > 0 {
			v.rolesMap = map[string]bool{}
			for _, r := range v.Security.Roles {
				v.rolesMap[r] = true
			}
		}
	}

}

// NewManager - initialize a new manager
func NewManager(configpath string) (*Manager, error) {
	manager := new(Manager)
	manager.macros = make(map[string]*Macro)
	manager.compiled = template.New("main")

	loadedFileCount := 0

	for _, p := range strings.Split(configpath, ",") {
		files, _ := filepath.Glob(p)

		if len(files) < 1 {
			log.Printf("Not found files in path(%s)\n", p)
		}

		if *flagDebug > 2 {
			log.Printf("%s matchs %v \n", p, files)
		}

		for _, file := range files {

			if *flagDebug > 2 {
				log.Printf("parse hcl file %s\n", file)
			}

			data, err := ioutil.ReadFile(file)
			if err != nil {
				if *flagDebug > 0 {
					log.Printf("read file(%s) error: %v\n", file, err)
				}
				return nil, err
			}

			loadedFileCount++

			var config map[string]*Macro
			if err := hcl.Unmarshal(data, &config); err != nil {
				if *flagDebug > 0 {
					log.Printf("%s syntax error: %v\n==hcl==\n%s\n", file, err, data)
				}
				return nil, err
			}

			for k, v := range config {
				//定义文件
				v.file = file
				if k == "_meta" {
					if manager.meta == nil {
						v.name = "_meta"
						manager.meta = v
						if manager.meta.Const != nil {
							manager.meta.consts, err = v.buildConst()
							if err != nil {
								return nil, err
							}
						}
					} else if *flagDebug > 0 {
						log.Printf("%s was already defined in %s, in %s duplication\n",
							v.name, manager.meta.file, v.file)
					}
					continue
				}

				manager.macros[k] = v

				if len(v.Provider) > 0 {
					_, err := manager.compiled.New(k).Parse(v.Provider)
					if err != nil {
						if *flagDebug > 0 {
							log.Printf("%s syntax error at %s provider: %v\n==provider==\n%s\n", file, k, err, v.Provider)
						}
						return nil, err
					}
				}

				if len(v.Exec) > 0 {
					_, err := manager.compiled.New(k).Parse(v.Exec)
					if err != nil {
						if *flagDebug > 0 {
							log.Printf("%s syntax error at %s exec: %v\n==exec==\n%s\n", file, k, err, v.Exec)
						}
						return nil, err
					}
				}

				if len(v.Total) > 0 {
					_, err := manager.compiled.New(k + "Total").Parse(v.Total)
					if err != nil {
						return nil, err
					}
				}

				v.methodMacros = make(map[string]*Macro)

				if v.Get != nil {
					v.methodMacros["GET"] = v.Get
				}

				if v.Post != nil {
					v.methodMacros["POST"] = v.Post
				}

				if v.Put != nil {
					v.methodMacros["PUT"] = v.Put
				}

				if v.Patch != nil {
					v.methodMacros["PATCH"] = v.Patch
				}

				if v.Delete != nil {
					v.methodMacros["DELETE"] = v.Delete
				}

				v.manager = manager
				v.name = k

				if v.Path == "" {
					v.Path = "/" + v.name
				}

				if !strings.HasPrefix(v.Path, "/") {
					v.Path = "/" + v.Path
				}

				fixMacro(v)

				if v.Const != nil {
					v.consts, err = v.buildConst()
					if err != nil {
						return nil, err
					}
				}

				for k, childm := range v.methodMacros {
					childm.manager = manager
					childm.Methods = []string{k}
					childm.name = v.name + "." + strings.ToLower(k)
					childm.Path = v.Path

					if len(childm.Provider) > 0 {
						_, err := manager.compiled.New(k).Parse(childm.Provider)
						if err != nil {
							if *flagDebug > 0 {
								log.Printf("%s syntax error at %s provider: %v\n==provider==\n%s\n",
									file, childm.name, err, childm.Provider)
							}
							return nil, err
						}
					}

					if len(childm.Exec) > 0 {
						_, err := manager.compiled.New(k).Parse(childm.Exec)
						if err != nil {
							if *flagDebug > 0 {
								log.Printf("%s syntax error at %s exec: %v\n==exec==\n%s\n",
									file, childm.name, err, childm.Exec)
							}
							return nil, err
						}
					}

					if len(childm.Total) > 0 {
						_, err := manager.compiled.New(childm.name + "Total").Parse(childm.Total)
						if err != nil {
							return nil, err
						}
					}

					if childm.Include == nil {
						childm.Include = v.Include
					} else if v.Include != nil {
						childm.Include = append(childm.Include, v.Include[0:]...)
					}
					if childm.Tags == nil {
						childm.Tags = v.Tags
					} else if v.Tags != nil {
						childm.Tags = append(childm.Tags, v.Tags[0:]...)
					}

					if childm.Security == nil {
						childm.Security = v.Security
					}

					if childm.Proxy == nil {
						childm.Proxy = v.Proxy
					}

					fixMacro(childm)

					if childm.Const != nil {
						childm.consts, err = childm.buildConst()
						if err != nil {
							return nil, err
						}
					}
				}

			}
		}
	}

	if loadedFileCount == 0 {
		return nil, fmt.Errorf("not found any file in %s", *flagAPIFile)
	}

	if len(manager.macros) == 0 {
		return nil, fmt.Errorf("not found any macro in %s", *flagAPIFile)
	}

	if manager.meta == nil {
		manager.meta = new(Macro)
	}

	meta := manager.meta

	if meta.Debug > *flagDebug {
		*flagDebug = meta.Debug
	} else if meta.Debug < *flagDebug {
		meta.Debug = *flagDebug
	}

	if meta.Name == "" {
		meta.Name = *flagName
	}

	if meta.Desc == "" {
		meta.Desc = *flagDescription
	}

	if meta.Author == nil {
		meta.Author = new(Author)
	}

	if meta.Author.Name == "" {
		meta.Author.Name = *flagAuthor
	}

	if meta.Author.Email == "" {
		meta.Author.Email = *flagEmail
	}

	if meta.Version == "" {
		meta.Version = *flagVersion
	}

	if meta.Base == "" {
		meta.Base = *flagBasePath
	}

	if meta.Base != "/" && strings.HasSuffix(meta.Base, "/") {
		meta.Base = meta.Base[:len(meta.Base)-1]
	}

	if meta.Jwt == nil {
		meta.Jwt = new(JwtConfig)
	}

	if meta.Brand == "" {
		meta.Brand = serverBrand
	}

	if meta.Jwt.Rsa != "" {
		if meta.Jwt.Secret == "" {
			return nil, fmt.Errorf("%s jwt secret is empty in _meta", manager.meta.file)
		}
		tmpPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(manager.meta.Jwt.Rsa))
		if err != nil {
			return nil, fmt.Errorf("%s jwt rsa private error in _meta: %v", manager.meta.file, err)
		}

		meta.Jwt.privkey = tmpPrivateKey
		meta.Jwt.file = "<HCL内置>"
	} else if *flagRSAPrivkey != "" {
		meta.Jwt.Secret = *flagJWTSecret
		rsaKeyData, err := ioutil.ReadFile(*flagRSAPrivkey)
		if err != nil {
			return nil, fmt.Errorf("cann't read rsa private key file(%s)", *flagRSAPrivkey)
		}
		meta.Jwt.file = *flagRSAPrivkey
		tmpPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(rsaKeyData)
		if err != nil {
			return nil, fmt.Errorf("rsa private key file(%s) not pem format", *flagRSAPrivkey)
		}
		meta.Jwt.privkey = tmpPrivateKey

		if *flagJWTExpires > meta.Jwt.Expires {
			meta.Jwt.Expires = *flagJWTExpires
		}
	}

	if meta.Jwt.Expires < 10 {
		meta.Jwt.Expires = 1800
	}

	if meta.Proxy == nil && *flagTrustedProxy != "" {
		meta.Proxy = strings.Split(*flagTrustedProxy, ",")
	}

	if !meta.Swagger {
		meta.Swagger = *flagSwagger
	}

	if meta.Security == nil {
		meta.Security = new(SecurityConfig)
	}

	if meta.Security.API == "" {
		meta.Security.API = *flagUserAPI
	}

	if meta.Security.Scope == "" {
		meta.Security.Scope = *flagUserScope
	}

	if meta.Security.Idtype == "" {
		meta.Security.Idtype = *flagUserIDType
	}

	if meta.Db == nil {
		meta.Db = new(DatabaseConfig)
	}

	if meta.Db.Driver == "" {
		meta.Db.Driver = *flagDBDriver
	}

	if meta.Db.Dsn == "" {
		meta.Db.Dsn = *flagDBDSN
	}

	if meta.Db.Redis == "" {
		meta.Db.Redis = *flagRedisURL
	}

	if meta.Db.Redis != "" {
		err := meta.Db.ConnectToRedis()
		if err != nil {
			return nil, err
		}
	}

	if meta.Mq == nil {
		meta.Mq = new(MessageQueueConfig)
	}

	for _, macro := range manager.macros {
		if len(meta.consts) > 0 {
			if len(macro.consts) > 0 {
				mapValue := map[string]interface{}{}
				for k, v := range meta.consts {
					mapValue[k] = v
				}
				for k, v := range macro.consts {
					mapValue[k] = v
				}
				macro.consts = mapValue
			} else {
				macro.consts = meta.consts
			}
		}

		for _, childm := range macro.methodMacros {
			if len(macro.consts) > 0 {
				if len(childm.consts) > 0 {
					mapValue := map[string]interface{}{}
					for k, v := range macro.consts {
						mapValue[k] = v
					}
					for k, v := range childm.consts {
						mapValue[k] = v
					}
					childm.consts = mapValue
				} else {
					childm.consts = macro.consts
				}
			}
		}
	}

	if meta.Mq.IsMessageQueueEnabled() {
		for _, n := range manager.Names() {
			m := manager.Get(n)
			if m.IsMessageConsumeEnabled() {
				if err := meta.Mq.NewMessageQueueProvider(m); err != nil {
					return nil, err
				}
			}
		}
	}

	return manager, nil
}

// Get - fetches the required macro
func (m *Manager) Get(macro string) *Macro {
	m.RLock()
	defer m.RUnlock()

	return m.macros[macro]
}

// Size - return the size of the currently loaded configs
func (m *Manager) Size() int {
	return len(m.macros)
}

// Names - return a list of registered macros
func (m *Manager) Names() (ret []string) {
	for k := range m.macros {
		ret = append(ret, k)
	}
	return ret
}

// List - return all macros
func (m *Manager) List() (ret []*Macro) {
	m.RLock()
	defer m.RUnlock()
	for _, v := range m.macros {
		if !strings.HasPrefix(v.name, "_") {
			ret = append(ret, v)
		}
	}
	return ret
}

// ServiceName - return name of service
func (m *Manager) ServiceName() string {
	return m.meta.Name
}

// ServiceDesc - return desc of service
func (m *Manager) ServiceDesc() string {
	return m.meta.Desc
}

// ServiceVersion - return desc of service
func (m *Manager) ServiceVersion() string {
	return m.meta.Version
}

// ServiceAuthor - return author of service
func (m *Manager) ServiceAuthor() *Author {
	return m.meta.Author
}

// ServiceBasePath - return is swagger enabled
func (m *Manager) ServiceBasePath() string {
	return m.meta.Base
}

// IsSwaggerEnabled - return is swagger enabled
func (m *Manager) IsSwaggerEnabled() bool {
	return m.meta.Swagger
}

// TrustedProxyList - return is swagger enabled
func (m *Manager) TrustedProxyList() []string {
	return m.meta.Proxy
}

// DebugLevel - return debug level
func (m *Manager) DebugLevel() int {
	return m.meta.Debug
}

// JwtIdentityConfig - return jwt identity config
func (m *Manager) JwtIdentityConfig() *JwtConfig {
	return m.meta.Jwt
}

// SecurityConfig - return security config
func (m *Manager) SecurityConfig() *SecurityConfig {
	return m.meta.Security
}

// DatabaseConfig - return database config
func (m *Manager) DatabaseConfig() *DatabaseConfig {
	return m.meta.Db
}

//ServiceBrand - return service brand
func (m *Manager) ServiceBrand() string {
	return m.meta.Brand
}

//MessageQueueConfig - 获取消息队列配置
func (m *Manager) MessageQueueConfig() *MessageQueueConfig {
	return m.meta.Mq
}

//IsMessageQueueEnabled - 是否启用了消息队列管理
func (m *Manager) IsMessageQueueEnabled() bool {
	return m.MessageQueueConfig().IsMessageQueueEnabled()
}
