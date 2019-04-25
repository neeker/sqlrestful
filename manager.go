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
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/hashicorp/hcl"
)

// Manager - a macros manager
type Manager struct {
	macros   map[string]*Macro
	compiled *template.Template
	sync.RWMutex
}

func fixMacro(v *Macro) error {
	if len(v.Total) > 0 {
		v.Type = "page"
	}

	if v.Type == "" {
		v.Type = "list"
	} else {
		v.Type = strings.ToLower(v.Type)
	}

	if v.Impl == "" {
		v.Impl = "sql"
	} else {
		v.Impl = strings.ToLower(v.Impl)
	}

	if v.Impl != "js" {
		v.Impl = "sql"
	}

	if v.Ret == "" {
		v.Ret = "enclosed"
	} else {
		v.Ret = strings.ToLower(v.Ret)
	}

	if v.Ret != "origin" {
		v.Ret = "enclosed"
	}

	switch {
	case v.Type == "list":
	case v.Type == "object":
	case v.Type == "page":
	default:
		v.Type = "list"
	}

	if v.Summary == "" {
		v.Summary = v.name
	}

	if len(v.Tags) == 0 {
		v.Tags = append(v.Tags, v.name)
	}

	return nil
}

// NewManager - initialize a new manager
func NewManager(configpath string) (*Manager, error) {
	manager := new(Manager)
	manager.macros = make(map[string]*Macro)
	manager.compiled = template.New("main")

	for _, p := range strings.Split(configpath, ",") {
		files, _ := filepath.Glob(p)

		if len(files) < 1 {
			return nil, fmt.Errorf("无效的配置文件路径 (%s)", p)
		}

		for _, file := range files {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				return nil, err
			}

			var config map[string]*Macro
			if err := hcl.Unmarshal(data, &config); err != nil {
				return nil, err
			}

			for k, v := range config {
				manager.macros[k] = v

				if len(v.Exec) > 0 {
					_, err := manager.compiled.New(k).Parse(v.Exec)
					if err != nil {
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

				for k, childm := range v.methodMacros {
					childm.manager = manager
					childm.Methods = []string{k}
					childm.name = v.name + "." + strings.ToLower(k)
					childm.Path = v.Path
					if childm.Tags == nil {
						childm.Tags = v.Tags
					} else {
						for _, t := range v.Tags {
							childm.Tags = append(childm.Tags, t)
						}
					}
					fixMacro(childm)
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

// List - return a list of registered macros
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
